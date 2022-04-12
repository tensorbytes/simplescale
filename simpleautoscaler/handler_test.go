package simpleautoscaler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tensorbytes/simplescale/utils"
	"github.com/tidwall/gjson"
	klogv2 "k8s.io/klog/v2"
)

func TestGjsonMatch(t *testing.T) {
	// study gjson Path Syntax
	jsonByte := `{
		"spec": {
			"containers": [
				{
					"name": "test"
					"resources": {
						"limits": {
							"cpu": 1
						}
					}
				}
			]
		}
	}`
	result := gjson.Parse(jsonByte)
	assert.Equal(t, result.Get(`spec.containers.#(name="test").resources.limits.cpu`).Raw, "1", "check error wapper")
	assert.Equal(t, result.Get(`spec.containers.0.resources.limits.cpu`).Raw, "1", "check error wapper")
}

func TestGetFunc(t *testing.T) {
	kubeClient, err := utils.GetControllerClient()
	if err != nil {
		panic(err)
	}
	uhandler := NewUnstructruedHandler(context.TODO(), kubeClient)
	r, err := uhandler.GetResource("VirtualService", "networking.istio.io/v1beta1", "redis-test", "default")
	if err != nil {
		panic(err)
	}
	klogv2.Info(r.Obj.GetName())
	result, ok := r.Query("spec.gateways.[0]")
	if ok {
		klogv2.Info(result)
	}
	lables := map[string]string{
		"app.oam.dev/component": "redis-test",
	}
	rlist, err := uhandler.ListResourceByLabels("VirtualService", "networking.istio.io/v1beta1", "default", lables)
	if err != nil {
		panic(err)
	}
	klogv2.Info(rlist.Objlist.Items)
	lables = make(map[string]string)
	rlist, err = uhandler.ListResourceByLabels("VirtualService", "networking.istio.io/v1beta1", "default", lables)
	if err != nil {
		panic(err)
	}
	klogv2.Info(rlist)
}

func TestUpdateFunc(t *testing.T) {
	kubeClient, err := utils.GetControllerClient()
	if err != nil {
		panic(err)
	}
	uhandler := NewUnstructruedHandler(context.TODO(), kubeClient)
	// test int field
	fieldValue := map[string]string{"spec.maxReplicas": "3"}
	err = uhandler.UpdateResourceFieldValue("HorizontalPodAutoscaler", "autoscaling/v1", "autoscale-test", "default", fieldValue)
	if err != nil {
		klogv2.Error(err)
	}
	// test quantity field
	fieldValue = map[string]string{`spec.template.spec.containers.#(name="autoscale-test").resources.limits.cpu`: "200m"}
	err = uhandler.UpdateResourceFieldValue("Deployment", "apps/v1", "autoscale-test", "default", fieldValue)
	if err != nil {
		klogv2.Error(err)
	}
}
