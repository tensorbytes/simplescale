package tests

import (
	"context"
	"fmt"
	"log"
	"time"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	"github.com/tensorbytes/simplescale/utils"
	"github.com/tidwall/gjson"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewTestEnv() (e *Env) {
	ctrClient, err := utils.GetControllerClient()
	if err != nil {
		panic(err)
	}
	e = &Env{
		kubeclient: ctrClient,
		obj:        unstructured.Unstructured{},
	}
	return
}

type Env struct {
	kubeclient       runtimeclient.Client
	obj              unstructured.Unstructured
	installResources []runtimeclient.Object
}

func (e *Env) PrepareResource(resources []string) {
	for _, resource := range resources {
		err := e.obj.UnmarshalJSON([]byte(resource))
		if err != nil {
			panic(err)
		}
		err = e.kubeclient.Create(context.TODO(), &e.obj, &runtimeclient.CreateOptions{})
		if err != nil {
			panic(err)
		}
		e.installResources = append(e.installResources, &e.obj)
	}
	time.Sleep(1 * time.Second)
}

func (e *Env) Clean() {
	time.Sleep(1 * time.Second)
	var errs []error
	for _, obj := range e.installResources {
		err := e.kubeclient.Delete(context.TODO(), obj, &runtimeclient.DeleteOptions{})
		errs = append(errs, err)
	}
	for _, e := range errs {
		if e != nil {
			panic(e)
		}
	}
}

func (e *Env) CreateDeployment(name, cpu, memory string, replicas int32) {
	cpuQuantity := apiresource.MustParse(cpu)
	memoryQuantity := apiresource.MustParse(cpu)
	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    name,
							Image:   "registry.cn-beijing.aliyuncs.com/tensorbytes/busybox:1.35.0-random-metrics",
							Command: []string{"/app/random-metrics"},
							Resources: corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]apiresource.Quantity{
									"cpu":    cpuQuantity,
									"memory": memoryQuantity,
								},
								Limits: map[corev1.ResourceName]apiresource.Quantity{
									"cpu":    cpuQuantity,
									"memory": memoryQuantity,
								},
							},
						},
					},
				},
			},
		},
	}
	err := e.kubeclient.Create(context.TODO(), &deployment, &runtimeclient.CreateOptions{})

	if err != nil && (!kubeerrors.IsAlreadyExists(err)) {
		panic(err)
	}
	e.installResources = append(e.installResources, &deployment)
	time.Sleep(1 * time.Second)
}

func (e *Env) CreateAutoScaler(name, targetRefName, targetRefField, scaleFactorName, scaleFactorNumber string) {
	simpleAutoScaler := autoscalev1alpha1.SimpleAutoScaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: autoscalev1alpha1.SimpleAutoScalerSpec{
			TargetRef: &autoscalev1alpha1.ScaleTargetResourceReference{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
				Name:       targetRefName,
			},
			Policy: []*autoscalev1alpha1.ScaleResourcePolicy{
				{
					Name:      "cpu-limit",
					Field:     targetRefField,
					FieldType: "cpu",
					Update: &autoscalev1alpha1.ScaleUpdateParameter{
						DownscaleWindow: "1m",
						UpscaleWindow:   "1m",
					},
					ScaleFactorObject: &autoscalev1alpha1.ResourceScaleFactor{
						Kind:       "RecommendationScaleFactor",
						APIVersion: "autoscale.scale.shikanon.com/v1alpha1",
						Name:       scaleFactorName,
						Field:      "status.scaleFactor",
					},
				},
			},
		},
	}
	scaleFactor := autoscalev1alpha1.RecommendationScaleFactor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      scaleFactorName,
			Namespace: "default",
		},
		Spec: autoscalev1alpha1.RecommendationScaleFactorSpec{
			Ref: &autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
				Name:       targetRefName,
			},
			Query: "",
		},
		Status: autoscalev1alpha1.RecommendationScaleFactorStatus{
			ScaleFactor: scaleFactorNumber,
		},
	}
	err := e.kubeclient.Create(context.TODO(), &scaleFactor, &runtimeclient.CreateOptions{})
	if err != nil {
		panic(err)
	}
	err = e.kubeclient.Status().Patch(context.TODO(), &scaleFactor, runtimeclient.Merge, &runtimeclient.PatchOptions{})
	if err != nil {
		panic(err)
	}
	e.installResources = append(e.installResources, &scaleFactor)
	err = e.kubeclient.Create(context.TODO(), &simpleAutoScaler, &runtimeclient.CreateOptions{})
	if err != nil {
		panic(err)
	}
	e.installResources = append(e.installResources, &simpleAutoScaler)
	time.Sleep(1 * time.Second)
}

// assert the resource field path equeue expect value
// 检查资源的字段是否等于期望的目标值
func (e *Env) Assert(kind, apiVersion, name, namespace, path, value string) {
	// waiting some time to check the changing
	time.Sleep(1 * time.Second)
	e.obj.SetAPIVersion(apiVersion)
	e.obj.SetKind(kind)
	searchkey := runtimeclient.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}
	err := e.kubeclient.Get(context.TODO(), searchkey, &e.obj)
	if err != nil {
		err = fmt.Errorf("%w,kind:%s,apiVersion:%s,name:%s,namespace:%s,path:%s,expectValue:%s", err, kind, apiVersion, name, namespace, path, value)
		log.Fatal(err)
	}
	resourceJson, err := e.obj.MarshalJSON()
	if err != nil {
		err = fmt.Errorf("%w,kind:%s,apiVersion:%s,name:%s,namespace:%s,path:%s,expectValue:%s", err, kind, apiVersion, name, namespace, path, value)
		log.Fatal(err)
	}
	fieldValue := gjson.ParseBytes(resourceJson).Get(path)
	if fieldValue.String() != value {
		err = fmt.Errorf("fieldValue(%s) no equal expectValue,kind:%s,apiVersion:%s,name:%s,namespace:%s,path:%s,expectValue:%s", fieldValue.String(), kind, apiVersion, name, namespace, path, value)
		log.Fatal(err)
	}
}
