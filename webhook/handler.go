package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	"github.com/tensorbytes/simplescale/utils"

	"github.com/tidwall/sjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	klogv2 "k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	kubeadmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// MutatingHandler handles Component
type ScaleMutatingHandler struct {
	Client runtimeclient.Client

	// Decoder decodes objects
	Decoder *kubeadmission.Decoder

	// cache for the resource
	CacheResource *sync.Map

	Ctx context.Context
	// stop command
	stopCh chan int
}

func NewScaleMutatingHandler() (h *ScaleMutatingHandler, err error) {
	ctrlClient, err := utils.GetControllerClient()
	if err != nil {
		return
	}
	scheme := kuberuntime.NewScheme()
	decoder, err := kubeadmission.NewDecoder(scheme)
	if err != nil {
		return
	}
	ctx := context.Background()
	h = &ScaleMutatingHandler{
		Client:        ctrlClient,
		Ctx:           ctx,
		Decoder:       decoder,
		CacheResource: &sync.Map{},
	}
	return
}

// Handle handles admission requests.
func (h *ScaleMutatingHandler) Handle(ctx context.Context, req kubeadmission.Request) kubeadmission.Response {
	newobj := &unstructured.Unstructured{}
	// reqJson, err := json.Marshal(req.AdmissionRequest)
	// if err != nil {
	// 	return kubeadmission.Errored(http.StatusBadRequest, err)
	// }
	// klogv2.Info("Request: ", string(reqJson))
	err := h.Decoder.Decode(req, newobj)
	if err != nil {
		return kubeadmission.Errored(http.StatusBadRequest, err)
	}
	// get object kind, name, namespace
	objKind := req.AdmissionRequest.RequestKind.Kind
	// objAPIServer := newobj.GetResourceVersion()
	objName := newobj.GetName()
	objNamespace := newobj.GetNamespace()
	// get cache
	CacheResourceName := fmt.Sprintf("%s:%s:%s", objKind, objName, objNamespace)

	Cachevalue, ok := h.CacheResource.Load(CacheResourceName)
	// 不在缓存中直接让更新通过
	if !ok {
		return kubeadmission.Allowed("yes")
	}
	scaleResource := Cachevalue.(autoscalev1alpha1.SimpleAutoScalerResources)
	klogv2.Info("resource name: ", CacheResourceName)

	oldObjJson, err := json.Marshal(newobj)
	if err != nil {
		klogv2.Error(err)
		return kubeadmission.Allowed("yes")
	}
	// base scaleResource fix and update the resource filed
	updateObjJson := oldObjJson
	for _, field := range scaleResource.ResourceFields {
		// Resource field 未初始化
		if field.DesiredFieldValue.String() == "" || field.Path == "" {
			return kubeadmission.Allowed("yes")
		}
		// set desired value to old obj
		updateObjJson, err = sjson.SetBytes(oldObjJson, field.Path, field.DesiredFieldValue.String())
		if err != nil {
			klogv2.Error(err)
			return kubeadmission.Allowed("yes")
		}
	}

	resp := kubeadmission.PatchResponseFromRaw(oldObjJson, updateObjJson)

	// patchs := make([]jsonpatchv2.Operation, 0)
	// for _, field := range scaleResource.ResourceFields {
	// 	path := field.Path
	// 	op := "replace"
	// 	value := field.DesiredFieldValue.String()
	// 	klogv2.Infof("jsonPatch: %s, %s, %s", path, op, value)
	// 	patchs = append(patchs, jsonpatchv2.NewOperation(op, path, value))
	// }

	// resp := kubeadmission.Patched("keep for the simpleautoscaler", patchs...)

	klogv2.Info("Response: ", resp.String())
	return resp
}

func (h *ScaleMutatingHandler) Sync() {
	for {
		select {
		case <-h.stopCh:
			return
		default:
			var simpleAutoScalerList autoscalev1alpha1.SimpleAutoScalerList
			err := h.Client.List(h.Ctx, &simpleAutoScalerList, &runtimeclient.ListOptions{})
			if err != nil {
				klogv2.Error(err)
			}
			for _, autoScalerItem := range simpleAutoScalerList.Items {
				for _, scaleResource := range autoScalerItem.Status.Resources {
					name := fmt.Sprintf("%s:%s:%s", scaleResource.Target.Kind, scaleResource.Target.Name, autoScalerItem.Namespace)
					h.CacheResource.Store(name, *scaleResource)
				}
			}
		}
	}

}

func (h *ScaleMutatingHandler) Close() {
	h.stopCh <- 1
	h.Ctx.Done()
}
