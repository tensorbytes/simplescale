package utils

import (
	"context"
	"fmt"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeClient struct {
	Clientset *kubefake.Clientset
}

// fake list
func (fc *FakeClient) AddReactorListRecommendationScaleFactor(obj runtime.Object) {
	fc.Clientset.AddReactor("list", "recommendationscalefactors", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, obj, nil
	})
}

// fake get
func (fc *FakeClient) AddReactorGetRecommendationScaleFactor(obj runtime.Object) {
	recommendationScaleFactorList, ok := obj.(*autoscalev1alpha1.RecommendationScaleFactorList)
	fc.Clientset.AddReactor("get", "recommendationscalefactors", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		name := action.(kubetesting.GetAction).GetName()
		if ok {
			for _, item := range recommendationScaleFactorList.Items {
				if item.Name == name {
					return true, &item, nil
				}
			}
		}
		return false, nil, fmt.Errorf("could not find the requested recommendationscalefactors: %s", name)
	})
}

// fake update
func (fc *FakeClient) AddReactorUpdateRecommendationScaleFactor() {
	fc.Clientset.AddReactor("update", "recommendationscalefactors", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		obj := action.(kubetesting.UpdateAction).GetObject().(*autoscalev1alpha1.RecommendationScaleFactor)
		return true, obj, nil
	})
}

// client.Client interface
func (fc *FakeClient) Create(ctx context.Context, obj runtimeclient.Object, opts ...runtimeclient.CreateOption) error {
	return nil

}

func NewTestRecommendationScaleFactorClient(objs autoscalev1alpha1.RecommendationScaleFactorList) *kubefake.Clientset {
	client := FakeClient{
		Clientset: &kubefake.Clientset{},
	}
	client.AddReactorListRecommendationScaleFactor(&objs)
	client.AddReactorGetRecommendationScaleFactor(&objs)
	client.AddReactorUpdateRecommendationScaleFactor()
	return client.Clientset
}
