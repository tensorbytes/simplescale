package simpleautoscaler

import (
	"crypto/md5"
	"errors"
	"fmt"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	DefaultDownscaleWindow = "5m"
	DefaultUpscaleWindow   = "5m"
	DefaultFieldType       = autoscalev1alpha1.OtherResourceType
)

func SimpleAutoScalerMutating(scaleObj *autoscalev1alpha1.SimpleAutoScaler) {
	for _, policy := range scaleObj.Spec.Policy {
		if policy.FieldType == "" {
			policy.FieldType = DefaultFieldType
		}
		if policy.Name == "" {
			policy.Name = fmt.Sprintf("%v", md5.Sum([]byte(policy.Field)))
		}
		if policy.Update == nil {
			policy.Update = &autoscalev1alpha1.ScaleUpdateParameter{
				DownscaleWindow: DefaultDownscaleWindow,
				UpscaleWindow:   DefaultUpscaleWindow,
				Mode:            autoscalev1alpha1.UpdateModeDefault,
			}
		}
		if policy.FieldType == autoscalev1alpha1.CPUResourceType {
			if policy.Update.MaxAllowed.String() == "0" {
				policy.Update.MaxAllowed = apiresource.MustParse("1")
			}
			if policy.Update.MinAllowed.String() == "0" {
				policy.Update.MinAllowed = apiresource.MustParse("10m")
			}
		} else if policy.FieldType == autoscalev1alpha1.MemoryResourceType {
			if policy.Update.MaxAllowed.String() == "0" {
				policy.Update.MaxAllowed = apiresource.MustParse("1Gi")
			}
			if policy.Update.MinAllowed.String() == "0" {
				policy.Update.MinAllowed = apiresource.MustParse("10Mi")
			}
		} else if (policy.FieldType == autoscalev1alpha1.ReplicasResourceType) || (policy.FieldType == DefaultFieldType) {
			if policy.Update.MaxAllowed.String() == "0" {
				policy.Update.MaxAllowed = apiresource.MustParse("2")
			}
			if policy.Update.MinAllowed.String() == "0" {
				policy.Update.MinAllowed = apiresource.MustParse("1")
			}
		}
		if policy.Update.DownscaleWindow == "" {
			policy.Update.DownscaleWindow = DefaultDownscaleWindow
		}
		if policy.Update.UpscaleWindow == "" {
			policy.Update.UpscaleWindow = DefaultUpscaleWindow
		}
		if policy.Update.Mode == "" {
			policy.Update.Mode = autoscalev1alpha1.UpdateModeDefault
		}
	}
}

func SimpleAutoScalerValidating(scaleObj autoscalev1alpha1.SimpleAutoScaler) (err error) {
	if scaleObj.Spec.TargetRef == nil {
		err = errors.New("SimpleAutoScaler TargetRef must have object not be nil")
		return
	}
	if (scaleObj.Spec.TargetRef.Kind == "") || (scaleObj.Spec.TargetRef.APIVersion == "") {
		err = errors.New("SimpleAutoScaler TargetRef kind or APIVersion must have object not be nil")
		return
	}
	if len(scaleObj.Spec.Policy) < 0 {
		err = errors.New("SimpleAutoScaler policy must have object not be nil")
		return
	}
	for _, policy := range scaleObj.Spec.Policy {
		if policy.Field == "" {
			err = errors.New("SimpleAutoScaler policy field must not be nil")
			return
		}
		if policy.ScaleFactorObject == nil {
			err = errors.New("SimpleAutoScaler policy ScaleFactorObject must not be nil")
			return
		}
		if (policy.ScaleFactorObject.Kind == "") || (policy.ScaleFactorObject.APIVersion == "") {
			err = errors.New("SimpleAutoScaler policy ScaleFactorObject kind or APIVersion must have object not be nil")
			return
		}
	}
	return
}
