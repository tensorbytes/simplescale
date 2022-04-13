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
	CacheResourceName := GenCacheKey(objKind, objName, objNamespace)

	Cachevalue, ok := h.CacheResource.Load(CacheResourceName)
	// 不在缓存中直接让更新通过
	if !ok {
		klogv2.Info("resource name: ", CacheResourceName, "is not in simpleAutoScaler reference cache")
		return kubeadmission.Allowed("yes")
	}
	scaleResource := Cachevalue.(WebhookResourceCacheItem)
	klogv2.Info("resource name: ", CacheResourceName, "resource filed is ", scaleResource.String())

	oldObjJson, err := json.Marshal(newobj)
	if err != nil {
		klogv2.Error(err, "json marshal error for admission request newobj")
		return kubeadmission.Allowed("yes")
	}
	// base scaleResource fix and update the resource filed
	updateObjJson := oldObjJson
	for _, field := range scaleResource.Resources.ResourceFields {
		// Resource field 未初始化
		if field.DesiredFieldValue.String() == "" || field.Path == "" {
			klogv2.Info("resource name: ", CacheResourceName, "the desired field value is null or field path is null")
			return kubeadmission.Allowed("yes")
		}
		// set desired value to old obj
		updateObjJson, err = sjson.SetBytes(oldObjJson, field.Path, field.DesiredFieldValue.String())
		if err != nil {
			klogv2.Error(fmt.Sprintf("the path (%s) and value (%s) set obj failed", field.Path, field.DesiredFieldValue.String()), err)
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

	klogv2.Info("Success Response: ", resp.String())
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
					name := GenCacheKey(scaleResource.Target.Kind, scaleResource.Target.Name, autoScalerItem.Namespace)
					cacheItem := WebhookResourceCacheItem{
						Name:                 name,
						SimpleAutoScalerName: autoScalerItem.Name,
						Resources:            scaleResource,
					}
					h.CacheResource.Store(name, cacheItem)
				}
			}
		}
	}

}

func (h *ScaleMutatingHandler) Close() {
	h.stopCh <- 1
	h.Ctx.Done()
}

// resource cache is use to filter requests
type WebhookResourceCacheItem struct {
	Name                 string
	SimpleAutoScalerName string
	Resources            autoscalev1alpha1.SimpleAutoScalerResources
}

func (w *WebhookResourceCacheItem) String() string {
	return fmt.Sprintf("cache item name: %s, SimpleAutoScaler name: %s, fields: %v", w.Name, w.SimpleAutoScalerName, w.Resources.ResourceFields)
}

func GenCacheKey(kind, name, namespace string) string {
	return fmt.Sprintf("[%s:%s:%s]", kind, name, namespace)
}
