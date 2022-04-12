package simpleautoscaler

import (
	"context"
	"testing"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	simplescaletests "github.com/tensorbytes/simplescale/tests"
	"github.com/tensorbytes/simplescale/utils"
)

// func TestResourceReader(t *testing.T) {
// 	kubeClient, err := utils.GetControllerClient()
// 	if err != nil {
// 		panic(err)
// 	}
// 	uhandler := NewUnstructruedHandler(context.TODO(), kubeClient)
// 	reader := ResourceManager{
// 		ResourceHandler: uhandler,
// 	}
// 	ns := ""
// 	policy := autoscalev1alpha1.ScaleResourcePolicy{
// 		Name:  "test",
// 		Field: "spec.replicas",
// 		ScaleFactorObject: &autoscalev1alpha1.ResourceScaleFactor{
// 			Kind:       "Deployment",
// 			APIVersion: "apps/v1",
// 			Name:       "redis-test",
// 			Namespace:  "default",
// 		},
// 	}
// 	reader.GetScaleFactorValue(context.TODO(), policy, ns)
// }

func TestListResourceReference(t *testing.T) {
	// prepare envirement
	e := simplescaletests.NewTestEnv()
	defer e.Clean()
	e.CreateDeployment("test", "50m", "100Mi", 1)
	// test
	kubeClient, err := utils.GetControllerClient()
	if err != nil {
		panic(err)
	}
	manager := NewResourceManager(kubeClient)
	target := autoscalev1alpha1.ScaleTargetResourceReference{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
		Name:       "test",
	}
	_, err = manager.ListResourceReference(context.TODO(), &target, "default")
	if err != nil {
		panic(err)
	}
}
