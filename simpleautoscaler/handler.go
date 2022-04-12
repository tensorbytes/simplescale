package simpleautoscaler

import (
	"context"
	"errors"

	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	kubefields "k8s.io/apimachinery/pkg/fields"
	kubelabels "k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrNotFoundPolicyField     = errors.New("not found policy field")
	ErrInvalidPolicyField      = errors.New("invalid policy field")
	ErrNotFoundResource        = errors.New("not found resource")
	ErrNotFoundResourceField   = errors.New("not found resource field")
	ErrUpdateResourceField     = errors.New("update resource field for error value")
	ErrUpdateResourceFieldType = errors.New("update resource field for error type")
)

type Handler interface {
	GetResource(kind, apiVersion, name, namespace string) (r Resource, err error)
	UpdateResourceFieldValue(kind, apiVersion, name, namespace string, fieldValue map[string]string) (err error)
	ListResourceByLabels(kind, apiVersion, namespace string, lablemap map[string]string) (r Resource, err error)
}

func NewUnstructruedHandler(ctx context.Context, kubeclient runtimeclient.Client) *UnstructruedHandler {
	return &UnstructruedHandler{
		ctx:    ctx,
		Client: kubeclient,
	}
}

type UnstructruedHandler struct {
	ctx    context.Context
	Client runtimeclient.Client
}

// get unstructrue
func (t *UnstructruedHandler) GetResource(kind, apiVersion, name, namespace string) (r Resource, err error) {
	r.Obj.SetAPIVersion(apiVersion)
	r.Obj.SetKind(kind)
	searchkey := runtimeclient.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}
	err = t.Client.Get(t.ctx, searchkey, &r.Obj)
	if err != nil {
		err = fmt.Errorf("%w: %v; resources name: %s,namespace: %s, kind: %s,apiVersion: %s,", ErrNotFoundResource, err, name, namespace, kind, apiVersion)
	}
	return
}

// update unstructrue field
func (t *UnstructruedHandler) UpdateResourceFieldValue(kind, apiVersion, name, namespace string, fieldValue map[string]string) (err error) {
	resource, err := t.GetResource(kind, apiVersion, name, namespace)
	if err != nil {
		return
	}
	resJson, err := resource.Obj.MarshalJSON()
	if err != nil {
		return
	}

	var updateResult []byte
	for field, value := range fieldValue {
		oldfieldValue := gjson.ParseBytes(resJson).Get(field)
		if !oldfieldValue.Exists() {
			// err is not found field
			err = fmt.Errorf("%w", ErrNotFoundResourceField)
			return err
		}
		// type check and convert
		var newfieldValue interface{}
		switch oldfieldValue.Type {
		case gjson.Number:
			newfieldValue, err = strconv.ParseInt(value, 10, 64)
		case gjson.String:
			newfieldValue = value
		default:
			err = fmt.Errorf("error field type %s", value)
		}
		if err != nil {
			return err
		}
		// update json bytes by field
		updateResult, err = sjson.SetBytes(resJson, field, newfieldValue)
		if err != nil {
			return err
		}
	}

	// new json struct load in object
	err = resource.Obj.UnmarshalJSON(updateResult)
	if err != nil {
		return err
	}
	// patch this new object
	err = t.Client.Patch(t.ctx, &resource.Obj, runtimeclient.Merge)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrUpdateResourceField, err)
	}
	return

}

func (t *UnstructruedHandler) RawUpdateResourceFieldValue(kind, apiVersion, name, namespace string, fieldValue map[string]string) (resource Resource, err error) {
	resource, err = t.GetResource(kind, apiVersion, name, namespace)
	if err != nil {
		return
	}
	resJson, err := resource.Obj.MarshalJSON()
	if err != nil {
		return
	}

	var updateResult []byte
	for field, value := range fieldValue {
		oldfieldValue := gjson.ParseBytes(resJson).Get(field)
		if !oldfieldValue.Exists() {
			// err is not found field
			err = fmt.Errorf("%w", ErrNotFoundResourceField)
			return resource, err
		}
		// type check and convert
		var newfieldValue interface{}
		switch oldfieldValue.Type {
		case gjson.Number:
			newfieldValue, err = strconv.ParseInt(value, 10, 64)
		case gjson.String:
			newfieldValue = value
		default:
			err = fmt.Errorf("error field type %s", value)
		}
		if err != nil {
			return resource, err
		}
		// update json bytes by field
		updateResult, err = sjson.SetBytes(resJson, field, newfieldValue)
		if err != nil {
			return resource, err
		}
	}

	// new json struct load in object
	err = resource.Obj.UnmarshalJSON(updateResult)
	return

}

// list unstructrue
func (t *UnstructruedHandler) ListResourceByLabels(kind, apiVersion, namespace string, lablemap map[string]string) (r Resource, err error) {
	r.Objlist.SetAPIVersion(apiVersion)
	r.Objlist.SetKind(kind)
	// the list function used field selector filter name, namespace
	// and use label selector filter labels
	labelSelector := kubelabels.SelectorFromSet(lablemap)
	fieldmap := map[string]string{
		"metadata.namespace": namespace,
	}
	fieldSelector := kubefields.SelectorFromSet(fieldmap)
	err = t.Client.List(t.ctx, &r.Objlist, &runtimeclient.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrNotFoundResource, err)
	}
	return
}
