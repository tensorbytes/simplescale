package simpleautoscaler

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	simplescaletests "github.com/tensorbytes/simplescale/tests"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// 测试获取目标资源
func TestGetTargetResource(t *testing.T) {
	// env
	e := simplescaletests.NewTestEnv()
	defer e.Clean()
	e.CreateDeployment("test", "50m", "100Mi", 1)
	e.CreateAutoScaler("test-autoscaler", "test", "spec.template.spec.containers.0.resources.limits.cpu", "test-scalefactor", "2")
	//
	autoScaler, err := NewSimpleAutoScalerController()
	if err != nil {
		log.Fatal(errors.Unwrap(err))
	}
	time.Sleep(2 * time.Second)
	var simpleAutoScaler autoscalev1alpha1.SimpleAutoScaler
	err = autoScaler.kubeClient.Get(context.TODO(), runtimeclient.ObjectKey{Name: "test-autoscaler", Namespace: "default"}, &simpleAutoScaler)
	if err != nil {
		log.Fatal(errors.Unwrap(err))
	}
	resources, err := autoScaler.GetTargetResource(context.TODO(), &simpleAutoScaler)
	if err != nil {
		log.Fatal(errors.Unwrap(err))
	}
	if resources[0].Target.Name != "test" {
		log.Fatal("target name no equals")
	}
}

// 测试策略更新,测试冷却时间
func TestUpdatePolicyResult(t *testing.T) {
	autoScaler, err := NewSimpleAutoScalerController()
	if err != nil {
		log.Fatal(errors.Unwrap(err))
	}
	now := time.Now()
	beforeOneMinute := now.Add(-1 * time.Minute)
	beforeTwoMinute := now.Add(-2 * time.Minute)
	beforeThreeMinute := now.Add(-3 * time.Minute)
	fakeSimpleAutoScaler := autoscalev1alpha1.SimpleAutoScaler{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cooldown",
		},
		Status: autoscalev1alpha1.SimpleAutoScalerStatus{
			Resources: []autoscalev1alpha1.SimpleAutoScalerResources{
				{
					ResourceFields: []autoscalev1alpha1.ResourceStautsField{
						{
							Name:              "cpu-requestes",
							CurrentFieldValue: apiresource.MustParse("100m"),
							LastUpScaleTime:   metav1.NewTime(beforeOneMinute),
							LastDownScaleTime: metav1.NewTime(beforeOneMinute),
						},
						{
							Name:              "cpu-limits",
							CurrentFieldValue: apiresource.MustParse("100m"),
							LastUpScaleTime:   metav1.NewTime(beforeThreeMinute),
							LastDownScaleTime: metav1.NewTime(beforeThreeMinute),
						},
						{
							Name:              "memory-requestes",
							CurrentFieldValue: apiresource.MustParse("100Mi"),
							LastUpScaleTime:   metav1.NewTime(beforeOneMinute),
							LastDownScaleTime: metav1.NewTime(beforeOneMinute),
						},
						{
							Name:              "memory-limits",
							CurrentFieldValue: apiresource.MustParse("100Mi"),
							LastUpScaleTime:   metav1.NewTime(beforeTwoMinute),
							LastDownScaleTime: metav1.NewTime(beforeTwoMinute),
						},
					},
				},
			},
		},
	}
	fakePolicyResultMap := map[string]PolicyResult{
		"cpu-requestes": {
			ScaleFactor:     2.0,
			UpscaleWindow:   2 * time.Minute,
			DownscaleWindow: 2 * time.Minute,
			MinAllowed:      apiresource.MustParse("1m"),
			MaxAllowed:      apiresource.MustParse("1000m"),
			Type:            "cpu",
		},
		"cpu-limits": {
			ScaleFactor:     0.5,
			UpscaleWindow:   2 * time.Minute,
			DownscaleWindow: 2 * time.Minute,
			MinAllowed:      apiresource.MustParse("1m"),
			MaxAllowed:      apiresource.MustParse("1000m"),
			Type:            "cpu",
		},
		"memory-requestes": {
			ScaleFactor:     2.0,
			UpscaleWindow:   2 * time.Minute,
			DownscaleWindow: 2 * time.Minute,
			MinAllowed:      apiresource.MustParse("1Mi"),
			MaxAllowed:      apiresource.MustParse("10Gi"),
			Type:            "memory",
		},
		"memory-limits": {
			ScaleFactor:     0.5,
			UpscaleWindow:   2 * time.Minute,
			DownscaleWindow: 2 * time.Minute,
			MinAllowed:      apiresource.MustParse("1Mi"),
			MaxAllowed:      apiresource.MustParse("10Gi"),
			Type:            "memory",
		},
	}
	autoScaler.UpdatePolicyResult(&fakeSimpleAutoScaler, fakePolicyResultMap)
	for _, r := range fakeSimpleAutoScaler.Status.Resources {
		for _, f := range r.ResourceFields {
			log.Println(f.DesiredFieldValue.String())
		}
	}
}

// test update function for DesiredFieldValue and Path
// 测试目标资源更新
func TestUpdateTargetResourceByDesired(t *testing.T) {
	e := simplescaletests.NewTestEnv()
	defer e.Clean()
	objName := "test-obj-update"
	e.CreateDeployment(objName, "50m", "100Mi", 1)
	e.Assert("Deployment", "apps/v1", objName, "default", "spec.template.spec.containers.0.resources.requests.cpu", "50m")
	ctrl, err := NewSimpleAutoScalerController()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	cpuQuantity := apiresource.MustParse("30m")
	simpleAutoScaler := autoscalev1alpha1.SimpleAutoScaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objName,
			Namespace: "default",
		},
		Spec: autoscalev1alpha1.SimpleAutoScalerSpec{
			TargetRef: &autoscalev1alpha1.ScaleTargetResourceReference{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
				Name:       objName,
			},
		},
		Status: autoscalev1alpha1.SimpleAutoScalerStatus{
			Resources: []autoscalev1alpha1.SimpleAutoScalerResources{
				{
					Target: &autoscalingv1.CrossVersionObjectReference{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
						Name:       objName,
					},
					ResourceFields: []autoscalev1alpha1.ResourceStautsField{
						{
							Name:              "cpu-requests",
							Path:              "spec.template.spec.containers.0.resources.requests.cpu",
							DesiredFieldValue: cpuQuantity,
						},
					},
				},
			},
		},
	}

	ctrl.UpdateTargetResourceByDesired(ctx, &simpleAutoScaler)
	e.Assert("Deployment", "apps/v1", objName, "default", "spec.template.spec.containers.0.resources.requests.cpu", "30m")
}

func TestRunOnce(t *testing.T) {
	e := simplescaletests.NewTestEnv()
	defer e.Clean()
	objDeploymentName := "test-run-once-deployment"
	e.CreateDeployment(objDeploymentName, "50m", "100Mi", 1)
	e.Assert("Deployment", "apps/v1", objDeploymentName, "default", "spec.template.spec.containers.0.resources.limits.cpu", "50m")
	e.CreateAutoScaler("test-autoscaler", "test-run-once-deployment", "spec.template.spec.containers.0.resources.limits.cpu", "test-scalefactor", "2")
	//
	autoScaler, err := NewSimpleAutoScalerController()
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)
	autoScaler.RunOnce(context.TODO())
	e.Assert("Deployment", "apps/v1", objDeploymentName, "default", "spec.template.spec.containers.0.resources.limits.cpu", "100m")
}
