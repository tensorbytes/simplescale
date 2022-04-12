package history

import (
	"context"
	"testing"
	"time"

	klogv2 "k8s.io/klog/v2"
)

// func init() {
// 	klogv2.InitFlags(nil)
// }

func NewFakePromethuesClient() {}

func getPrometheusHistoryProviderConfig() PrometheusHistoryProviderConfig {
	return PrometheusHistoryProviderConfig{
		Address:   "http://127.0.0.1:9090",
		QueryStep: 5 * time.Minute,
	}

}

func TestPromethuesQuery(t *testing.T) {
	config := getPrometheusHistoryProviderConfig()
	provider, err := NewPrometheusHistoryProvider(config)
	if err != nil {
		klogv2.Errorln(err)
		return
	}
	ctx := context.TODO()
	expression := `sum (rate (container_cpu_usage_seconds_total{
		image!="",container="autoscale-test",pod=~"^autoscale-test()().*$",kubernetes_io_hostname=~"^.*$"}[5m]
		)) by (pod)/sum (kube_pod_container_resource_requests_cpu_cores{pod=~"^autoscale-test()().*$"})by (pod)`
	val, err := provider.Query(ctx, expression)
	if err != nil {
		klogv2.Errorln(err)
		return
	}
	klogv2.Info(val.String())
	val, err = provider.QueryRange(ctx, expression, 1*time.Hour)
	if err != nil {
		klogv2.Errorln(err)
		return
	}
	klogv2.Info(val.String())
}
