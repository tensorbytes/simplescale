package simpleautoscaler

import (
	"context"
	"errors"
	"fmt"
	"time"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	"github.com/tensorbytes/simplescale/utils"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klogv2 "k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type AutoScalerController interface {
	RunOnce()
}

func NewSimpleAutoScalerController() (ctrl *SimpleAutoScalerController, err error) {
	ctrlClient, err := utils.GetControllerClient()
	if err != nil {
		return
	}
	mgr := NewResourceManager(ctrlClient)
	ctrl = &SimpleAutoScalerController{
		kubeClient: ctrlClient,
		Manager:    mgr,
	}
	return
}

type SimpleAutoScalerController struct {
	kubeClient runtimeclient.Client
	Manager    ResourceManager
}

func (s *SimpleAutoScalerController) RunOnce(ctx context.Context) {
	// for list to check every simpleautoscalerList
	var simpleAutoScalerList autoscalev1alpha1.SimpleAutoScalerList
	err := s.kubeClient.List(ctx, &simpleAutoScalerList, &runtimeclient.ListOptions{})
	if err != nil {
		klogv2.Errorf("list resource of SimpleAutoScaler error: %v", err)
		return
	}
	for _, autoScalerItem := range simpleAutoScalerList.Items {
		klogv2.Infof("handle simpleAutoScaler object: %v", autoScalerItem.Name)
		// webhook checking the simpleAutoScaler CRD
		err := SimpleAutoScalerValidating(autoScalerItem)
		if err != nil {
			klogv2.Error(err)
			break
		}
		SimpleAutoScalerMutating(&autoScalerItem)
		// update target resource. Get the target resource and
		// scale policy result, use policy result to update the target resource
		targetResources, err := s.GetTargetResource(ctx, &autoScalerItem)
		if err != nil {
			// not match resource
			autoScalerItem.Status.AddCondition(autoscalev1alpha1.MatchedAllResourceCondition, err)
			// update status
			go s.UpdateSimpleAutoScalerStatus(ctx, autoScalerItem)
			break
		}
		if len(targetResources) > 0 {
			// not match resource
			autoScalerItem.Status.AddCondition(autoscalev1alpha1.MatchedAllResourceCondition, nil)
		}
		policyResult, err := s.GetPolicyResult(&autoScalerItem)
		if err != nil {
			klogv2.Error(err)
			// update status
			go s.UpdateSimpleAutoScalerStatus(ctx, autoScalerItem)
			break
		}
		err = s.UpdatePolicyResult(&autoScalerItem, policyResult)
		if err != nil {
			klogv2.Error(err)
			// update status
			go s.UpdateSimpleAutoScalerStatus(ctx, autoScalerItem)
			break
		}
		klogv2.Infof("Policy Result:%v", policyResult)

		s.UpdateTargetResourceByDesired(ctx, &autoScalerItem)

		go s.UpdateSimpleAutoScalerStatus(ctx, autoScalerItem)
	}
}

type PolicyResultMap map[string]PolicyResult
type PolicyResult struct {
	Path            string
	ScaleFactor     float64
	MinAllowed      apiresource.Quantity
	MaxAllowed      apiresource.Quantity
	DownscaleWindow time.Duration
	UpscaleWindow   time.Duration
	Type            string
}

type SyncSpecFieldToStauts struct {
	Field string
}

// get target resource
// 获取待扩容的目标资源
func (s *SimpleAutoScalerController) GetTargetResource(ctx context.Context, scalerObj *autoscalev1alpha1.SimpleAutoScaler) (targetResource []*autoscalev1alpha1.SimpleAutoScalerResources, err error) {
	objReference, err := s.Manager.ListResourceReference(ctx, scalerObj.Spec.TargetRef, scalerObj.GetNamespace())
	if err != nil {
		return
	}
	specFieldMap := make(map[string]SyncSpecFieldToStauts)
	for _, specPolicyField := range scalerObj.Spec.Policy {
		specFieldMap[specPolicyField.Name] = SyncSpecFieldToStauts{
			Field: specPolicyField.Field,
		}
	}

	for _, obj := range objReference {
		newSimpleAutoScalerResources := autoscalev1alpha1.SimpleAutoScalerResources{
			Target:         obj,
			ResourceFields: make([]*autoscalev1alpha1.ResourceStautsField, 0),
		}

		// check replace or append
		resourceIsInStatus := false
		var replaceNumber int
		for i, resource := range scalerObj.Status.Resources {
			if resource.Target.Name == obj.Name {
				resourceIsInStatus = true
				replaceNumber = i
			}
		}
		if !resourceIsInStatus {
			// 不存在将所有的policy初始化一份到status
			// 如何将原来status 里面的 resource 资源赋值
			for _, specPolicyField := range scalerObj.Spec.Policy {
				statusField, err := s.addNewResourceStautsField(*specPolicyField, *obj, scalerObj.GetNamespace())
				if err != nil {
					klogv2.Error(err)
					break
				}
				newSimpleAutoScalerResources.ResourceFields = append(newSimpleAutoScalerResources.ResourceFields, &statusField)
			}
			scalerObj.Status.Resources = append(scalerObj.Status.Resources, &newSimpleAutoScalerResources)
		} else {
			// 存在需要保留原来的policy
			// replace origin resource
			newResourceFields := make([]*autoscalev1alpha1.ResourceStautsField, 0)
			oldSimpleAutoScalerResources := scalerObj.Status.Resources[replaceNumber]
			for _, specPolicyField := range scalerObj.Spec.Policy {
				notInStatusField := true
				for _, statusPolicyField := range oldSimpleAutoScalerResources.ResourceFields {
					if specPolicyField.Name == statusPolicyField.Name {
						// update currentFieldValue
						currentFieldValue, err := s.getCurrentFieldValue(specPolicyField.Field, *obj, scalerObj.Namespace)
						if err != nil {
							klogv2.Error(err)
							// jump to the loop
							goto ENDLOOP
						}
						statusPolicyField.CurrentFieldValue = currentFieldValue
						// set the old statusPolicyField to new
						newResourceFields = append(newResourceFields, statusPolicyField)
						notInStatusField = false
					}
				}
				if notInStatusField {
					statusField, err := s.addNewResourceStautsField(*specPolicyField, *obj, scalerObj.GetNamespace())
					if err != nil {
						klogv2.Error(err)
						break
					}
					newResourceFields = append(newResourceFields, &statusField)
				}
			ENDLOOP:
			}
			scalerObj.Status.Resources[replaceNumber].ResourceFields = newResourceFields
		}
	}
	targetResource = scalerObj.Status.Resources
	return
}

func (s *SimpleAutoScalerController) getCurrentFieldValue(field string,
	obj autoscalingv1.CrossVersionObjectReference, namespace string) (currentFieldValue apiresource.Quantity, err error) {
	r, err := s.Manager.ResourceHandler.GetResource(obj.Kind, obj.APIVersion, obj.Name, namespace)
	if err != nil {
		err = fmt.Errorf("ResourceManager cannot found resource: %w", err)
		return
	}
	result, ok := r.Query(field)
	if !ok || result == "" {
		err = fmt.Errorf("policy Field parse error: %s", field)
		return
	}
	currentFieldValue, err = apiresource.ParseQuantity(result)
	if err != nil {
		err = fmt.Errorf("value cannot parse to the quantity type, value is %s", result)
		return
	}
	return
}

// 识别到新资源，添加字段
func (s *SimpleAutoScalerController) addNewResourceStautsField(policy autoscalev1alpha1.ScaleResourcePolicy,
	obj autoscalingv1.CrossVersionObjectReference, namespace string) (statusField autoscalev1alpha1.ResourceStautsField, err error) {
	klogv2.Info("add new resource status field", policy.Field)
	currentFieldValue, err := s.getCurrentFieldValue(policy.Field, obj, namespace)
	if err != nil {
		klogv2.Errorf("value cannot parse to the quantity type, value is %s", currentFieldValue.String())
		return
	}
	upscaleWindow, err := time.ParseDuration(policy.Update.UpscaleWindow)
	if err != nil {
		return
	}
	downscaleWindow, err := time.ParseDuration(policy.Update.DownscaleWindow)
	if err != nil {
		return
	}
	statusField = autoscalev1alpha1.ResourceStautsField{
		Name:              policy.Name,
		Path:              policy.Field,
		LastUpScaleTime:   metav1.NewTime(time.Now().Add(-upscaleWindow)),
		LastDownScaleTime: metav1.NewTime(time.Now().Add(-downscaleWindow)),
		CurrentFieldValue: currentFieldValue,
	}
	return
}

// update target resource, use SimpleAutoScalerResources infomation to update target resource
// update the DesiredFieldValue to target resource field
// 根据期望值更新目标资源
func (s *SimpleAutoScalerController) UpdateTargetResourceByDesired(ctx context.Context, scalerObj *autoscalev1alpha1.SimpleAutoScaler) {
	fieldValue := make(map[string]string)
	for _, resource := range scalerObj.Status.Resources {
		for _, field := range resource.ResourceFields {
			if (field.DesiredFieldValue != apiresource.Quantity{}) {
				fieldValue[field.Path] = field.DesiredFieldValue.String()
			}
		}
	}

	resources := scalerObj.Status.Resources
	namespace := scalerObj.GetNamespace()
	if len(fieldValue) > 0 {
		for _, resource := range resources {
			err := s.Manager.UpdateScalerResources(*resource.Target, namespace, fieldValue)
			if err != nil {
				scalerObj.Status.AddCondition(autoscalev1alpha1.UpdateTargetResourceCondition, err)
				klogv2.Error(err)
				return
			}
		}
		scalerObj.Status.AddCondition(autoscalev1alpha1.UpdateTargetResourceCondition, nil)
	}

}

// 获取扩容策略结果
func (s *SimpleAutoScalerController) GetPolicyResult(scalerObj *autoscalev1alpha1.SimpleAutoScaler) (result PolicyResultMap, err error) {
	result = make(map[string]PolicyResult)
	for _, policy := range scalerObj.Spec.Policy {
		// get scale rate value
		value, err := s.Manager.GetScaleFactorValue(context.TODO(), *policy, scalerObj.GetNamespace())
		if err != nil {
			// policy filed
			if errors.Is(err, ErrNotFoundPolicyField) || errors.Is(err, ErrInvalidPolicyField) {
				scalerObj.Status.AddCondition(autoscalev1alpha1.CheckPolicyFieldValueCondition, err)
			}
			// get scale factor resource
			if errors.Is(err, ErrNotFoundResource) {
				scalerObj.Status.AddCondition(autoscalev1alpha1.GetScaleFactorCondition, err)
			}
			klogv2.Error(err)
			return result, err
		} else {
			scalerObj.Status.AddCondition(autoscalev1alpha1.CheckPolicyFieldValueCondition, nil)
			scalerObj.Status.AddCondition(autoscalev1alpha1.GetScaleFactorCondition, nil)
		}
		downscaleTime, err := time.ParseDuration(policy.Update.DownscaleWindow)
		if err != nil {
			scalerObj.Status.AddCondition(autoscalev1alpha1.UpdatedFieldCondition, err)
			return result, err
		}
		upscaleTime, err := time.ParseDuration(policy.Update.UpscaleWindow)
		if err != nil {
			scalerObj.Status.AddCondition(autoscalev1alpha1.UpdatedFieldCondition, err)
			return result, err
		}
		result[policy.Name] = PolicyResult{
			Path:            policy.Field,
			ScaleFactor:     value,
			MinAllowed:      policy.Update.MinAllowed,
			MaxAllowed:      policy.Update.MaxAllowed,
			DownscaleWindow: downscaleTime,
			UpscaleWindow:   upscaleTime,
		}
	}
	return
}

// 更新扩容策略结果
func (s *SimpleAutoScalerController) UpdatePolicyResult(scalerObj *autoscalev1alpha1.SimpleAutoScaler, result PolicyResultMap) (err error) {
	// handle cooldown error
	var endCooldownError error
	endCooldownErrorFormatFunc := func(name string, upOrDown bool, err error) error {
		var scaleAction string
		if upOrDown {
			scaleAction = "upscale"
		} else {
			scaleAction = "downscale"
		}
		if err == nil {
			return fmt.Errorf("resource is cooldown, can not be update, %s cannot %s", name, scaleAction)
		}
		return fmt.Errorf("%v, %s cannot %s", err, name, scaleAction)
	}
	for _, resource := range scalerObj.Status.Resources {
		// update scale rate value for status of resource
		for _, field := range resource.ResourceFields {
			if _, ok := result[field.Name]; ok {
				policyResult := result[field.Name]
				// 将不合法得scaleFactor排除在外
				if policyResult.ScaleFactor <= 0 {
					field.DesiredFieldValue = field.CurrentFieldValue
					continue
				}
				// 计算期望值，同时对比最大和最小边界
				desiredValue, err := utils.QuantityMultiplicative(field.CurrentFieldValue, policyResult.ScaleFactor)
				if err != nil {
					scalerObj.Status.AddCondition(autoscalev1alpha1.UpdatedFieldCondition, err)
					return fmt.Errorf("%w, parse desiredValue error", err)
				}

				if desiredValue.Cmp(policyResult.MinAllowed) < 0 {
					// desired value smaller than min value
					errorMessage := fmt.Errorf("desiredValue(%s) is smaller than min allowed value, so use min allowed instead of it", desiredValue.String())
					scalerObj.Status.AddCondition(autoscalev1alpha1.UpdatedFieldCondition, errorMessage)
					// set desiredvalue min allowed value
					desiredValue = policyResult.MinAllowed
				} else if desiredValue.Cmp(policyResult.MaxAllowed) > 0 {
					// desired value bigger than max value
					errorMessage := fmt.Errorf("desiredValue(%s) is bigger than max allowed value, so use max allowed instead of it", desiredValue.String())
					scalerObj.Status.AddCondition(autoscalev1alpha1.UpdatedFieldCondition, errorMessage)
					// set desiredvalue min allowed value
					desiredValue = policyResult.MaxAllowed
				} else {
					scalerObj.Status.AddCondition(autoscalev1alpha1.UpdatedFieldCondition, nil)
				}
				// 计算更新冷却时间
				timenow := time.Now()
				if policyResult.ScaleFactor > 1 {
					klogv2.Info(timenow)
					klogv2.Info(field.LastUpScaleTime)
					klogv2.Info(field.LastUpScaleTime.Add(policyResult.UpscaleWindow))
					if field.LastUpScaleTime.Add(policyResult.UpscaleWindow).Before(timenow) {
						field.LastUpScaleTime = metav1.NewTime(timenow)
					} else {
						endCooldownError = endCooldownErrorFormatFunc(field.Name, true, endCooldownError)
						continue
					}
				}
				if policyResult.ScaleFactor < 1 {
					klogv2.Info(timenow)
					klogv2.Info(field.LastDownScaleTime)
					klogv2.Info(field.LastDownScaleTime.Add(policyResult.DownscaleWindow))
					if field.LastDownScaleTime.Add(policyResult.DownscaleWindow).Before(timenow) {
						field.LastDownScaleTime = metav1.NewTime(timenow)
					} else {
						endCooldownError = endCooldownErrorFormatFunc(field.Name, false, endCooldownError)
						continue
					}
				}
				field.ScaleFactor = fmt.Sprintf("%f", policyResult.ScaleFactor)
				field.DesiredFieldValue = desiredValue
			}
		}
	}
	// if endCooldownError is nil, then it will true or it will set the endCooldownErrorFormatFunc result
	scalerObj.Status.AddCondition(autoscalev1alpha1.EndCooldownCondition, endCooldownError)
	return
}

// 真实执行更新 simpleautoscaler的 status
func (s *SimpleAutoScalerController) UpdateSimpleAutoScalerStatus(ctx context.Context, scalerObj autoscalev1alpha1.SimpleAutoScaler) {
	retry := 3
	err := s.updateSimpleAutoScalerStatus(ctx, &scalerObj, retry)
	if err != nil {
		klogv2.Errorf("%v, name:%s,namespace:%s", err, scalerObj.Name, scalerObj.Namespace)
	}
}

func (s *SimpleAutoScalerController) updateSimpleAutoScalerStatus(ctx context.Context, scalerObj *autoscalev1alpha1.SimpleAutoScaler, retry int) error {
	if err := s.kubeClient.Status().Update(ctx, scalerObj, &runtimeclient.UpdateOptions{}); err != nil {
		if apierrors.IsConflict(err) {
			retry = retry - 1
			if retry > 0 {
				err = s.updateSimpleAutoScalerStatus(ctx, scalerObj, retry)
			}
		}
		return fmt.Errorf("SimpleAutoScaler status update error: %v", err)
	}
	return nil
}
