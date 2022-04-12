package simpleautoscaler

import (
	"context"

	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
)

func NewResourceManager(kubeclient runtimeclient.Client) ResourceManager {
	ctx := context.Background()
	uhandler := NewUnstructruedHandler(ctx, kubeclient)
	return ResourceManager{
		ResourceHandler: uhandler,
	}
}

// resource reader for read value from simple autoscaler
type ResourceManager struct {
	ResourceHandler Handler
}

func (r *ResourceManager) ListResourceReference(ctx context.Context,
	target *autoscalev1alpha1.ScaleTargetResourceReference, namespace string) (cvor []*autoscalingv1.CrossVersionObjectReference, err error) {
	var resource Resource
	kind := target.Kind
	apiVersion := target.APIVersion
	name := target.Name
	if name != "" {
		resource, err = r.ResourceHandler.GetResource(kind, apiVersion, name, namespace)
		if err != nil {
			err = fmt.Errorf("%w, get resource error:%v", ErrNotFoundResource, err)
			return
		}
		cvor = resource.ToObjectReference()
		return
	} else {
		labels := target.Selector
		resource, err = r.ResourceHandler.ListResourceByLabels(kind, apiVersion, namespace, labels)
		if err != nil {
			err = fmt.Errorf("%w, list resource error:%v", ErrNotFoundResource, err)
			return
		}
		cvor = resource.ToObjectReference()
		return
	}
}

// update scaler resource function
func (r *ResourceManager) UpdateScalerResources(objReference autoscalingv1.CrossVersionObjectReference, namespace string, fieldValue map[string]string) (err error) {
	kind := objReference.Kind
	apiVersion := objReference.APIVersion
	name := objReference.Name
	err = r.ResourceHandler.UpdateResourceFieldValue(kind, apiVersion, name, namespace, fieldValue)
	if err != nil {
		err = fmt.Errorf("%w,name:%s,field:%v", err, name, fieldValue)
	}
	return
}

// query scale factor value though custom resources
func (r *ResourceManager) GetScaleFactorValue(ctx context.Context,
	policy autoscalev1alpha1.ScaleResourcePolicy, namespace string) (value float64, err error) {
	path := policy.ScaleFactorObject.Field
	kind := policy.ScaleFactorObject.Kind
	apiVersion := policy.ScaleFactorObject.APIVersion
	name := policy.ScaleFactorObject.Name
	value, err = r.ReadObjectByPath(ctx, kind, apiVersion, name, namespace, path)
	if err != nil {
		err = fmt.Errorf("%w, get scale factor error, kind: %s,name: %s, namespace: %s, path: %s,",
			err, kind, name, namespace, path)
	}
	return
}

func (r *ResourceManager) ReadObjectByPath(ctx context.Context,
	kind, apiVersion, name, namespace, path string) (value float64, err error) {
	resource, err := r.ResourceHandler.GetResource(kind, apiVersion, name, namespace)
	if err != nil {
		err = fmt.Errorf("ResourceManager cannot found resource: %w", err)
		return
	}
	result, ok := resource.Query(path)
	if ok {
		value, err = strconv.ParseFloat(result, 64)
		if err != nil {
			err = fmt.Errorf("%w: query invalid error in resource:%v", ErrInvalidPolicyField, err)
		}
	} else {
		err = fmt.Errorf("%w: not query the path in resource", ErrNotFoundPolicyField)
	}
	return
}

// use query
type Resource struct {
	Obj     unstructured.Unstructured
	Objlist unstructured.UnstructuredList
	Json    string
}

func (r *Resource) Query(path string) (value string, ok bool) {
	jsonobj, err := r.Obj.MarshalJSON()
	if err != nil {
		return
	}
	queryResult := gjson.GetBytes(jsonobj, path)
	value = queryResult.String()
	ok = true
	return
}

func (r *Resource) ToObjectReference() (references []*autoscalingv1.CrossVersionObjectReference) {
	if len(r.Objlist.Items) > 0 {
		for _, item := range r.Objlist.Items {
			var cvor autoscalingv1.CrossVersionObjectReference
			cvor.Kind = item.GetKind()
			cvor.APIVersion = item.GetAPIVersion()
			cvor.Name = item.GetName()
			references = append(references, &cvor)
		}
	}
	if (r.Obj.GetKind() != "") && (r.Obj.GetAPIVersion() != "") {
		var cvor autoscalingv1.CrossVersionObjectReference
		cvor.Kind = r.Obj.GetKind()
		cvor.APIVersion = r.Obj.GetAPIVersion()
		cvor.Name = r.Obj.GetName()
		references = append(references, &cvor)
	}
	return
}
