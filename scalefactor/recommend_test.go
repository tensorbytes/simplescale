package scalefactor

import (
	"context"
	"testing"
	"time"

	simplescalev1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	"github.com/tensorbytes/simplescale/scalefactor/history"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klogv2 "k8s.io/klog/v2"
)

func getRecommendationScaleFactor(name, query, value string) simplescalev1.RecommendationScaleFactor {
	return simplescalev1.RecommendationScaleFactor{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: simplescalev1.RecommendationScaleFactorSpec{
			Query:        query,
			DesiredValue: value,
		},
	}
}

func NewFakeSimpleRecommender() *SimpleRecommender {
	simpleRecommender, err := NewSimpleRecommender()
	if err != nil {
		panic(err)
	}
	provider, err := history.NewPrometheusHistoryProvider(history.PrometheusHistoryProviderConfig{
		Address:   "http://127.0.0.1:9090",
		QueryStep: 5 * time.Minute,
	})
	if err != nil {
		panic(err)
	}
	simpleRecommender.RegisterCollector(&provider)
	algorithm := NewSimpleScaleModel()
	simpleRecommender.RegisterAlgorithm(algorithm)
	return simpleRecommender
}

func TestGetRecommendResourceValue(t *testing.T) {
	query := `sum (rate (container_cpu_usage_seconds_total{
		image!="",container="autoscale-test",pod=~"^autoscale-test()().*$",kubernetes_io_hostname=~"^.*$"}[5m]
		)) by (pod)/sum (kube_pod_container_resource_requests_cpu_cores{pod=~"^autoscale-test()().*$"})by (pod)`
	testObj := getRecommendationScaleFactor("test-query", query, "0.02")

	simpleRecommender := NewFakeSimpleRecommender()
	queryResult, err := simpleRecommender.GetRecommendResourceValue(context.TODO(), testObj)
	if err != nil {
		klogv2.Error(err)
		return
	}
	klogv2.Info(queryResult)
}
