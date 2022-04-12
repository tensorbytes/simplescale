/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ScaleTargetResourceReference struct {
	// Kind of the reference resource
	Kind string `json:"kind"`
	// API version of the reference resource
	APIVersion string `json:"apiVersion,omitempty"`
	// Name of the referent
	Name string `json:"name,omitempty"`
	// Selector of match label
	Selector map[string]string `json:"selector,omitempty"`
}

type ScaleUpdateParameter struct {
	// the time interval of between downscale
	DownscaleWindow string `json:"downscaleWindow,omitempty"`
	// the time interval of between upscale
	UpscaleWindow string `json:"upscaleWindow,omitempty"`
	// the field range of minimum bound
	MinAllowed resource.Quantity `json:"minAllowed,omitempty"`

	// the field range of maximum bound
	MaxAllowed resource.Quantity `json:"maxAllowed,omitempty"`
	// update action mode
	Mode UpdateMode `json:"mode,omitempty"`
}

// UpdateMode controls whether can upscale or downscale
type UpdateMode string

const (
	// default is update can upscale and downsale
	UpdateModeDefault UpdateMode = "default"
	// onlyUpScale is update only for upscale
	UpdateModeOnlyUpscale UpdateMode = "onlyUpScale"
	// onlyDownScale is update only for upscale
	UpdateModeOnlyDownscale UpdateMode = "onlyDownScale"
)

type ScaleResourcePolicy struct {
	// the name of scalePolicy
	Name string `json:"name"`

	// the field of target resource will be updated
	Field string `json:"field"`

	// field type
	FieldType ScalePolicyResourceType `json:"type,omitempty"`

	// Update is parameter of update
	Update *ScaleUpdateParameter `json:"update,omitempty"`

	// description of the scale factor resource
	ScaleFactorObject *ResourceScaleFactor `json:"scaleFactorObject,omitempty"`
}

type ScalePolicyResourceType string

const (
	// cpu resource
	CPUResourceType ScalePolicyResourceType = "cpu"
	// memory resource
	MemoryResourceType ScalePolicyResourceType = "memory"
	// replicas resource
	ReplicasResourceType ScalePolicyResourceType = "replicas"
	// other extend resource, like gpu
	OtherResourceType ScalePolicyResourceType = "other"
)

type ResourceScaleFactor struct {
	// Kind of the reference resource
	Kind string `json:"kind"`

	// API version of the reference resource
	APIVersion string `json:"apiVersion,omitempty"`

	// Name of the referent
	Name string `json:"name,omitempty"`

	// Namespace
	Namespace string `json:"namespace,omitempty"`

	// the field of target resource will be updated
	Field string `json:"field,omitempty"`
}

// SimpleAutoScalerSpec defines the desired state of SimpleAutoScaler
type SimpleAutoScalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TargetRef is container scale resources
	TargetRef *ScaleTargetResourceReference `json:"targetRef"`

	// Policy is description rules to the relate parameter and update field
	Policy []*ScaleResourcePolicy `json:"policy,omitempty"`
}

// SimpleAutoScalerStatus defines the observed state of SimpleAutoScaler
type SimpleAutoScalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions is set of conditions required for autoscaler to scale target
	Conditions []SimpleAutoScalerCondition `json:"conditions"`

	// Resources which scaler select
	Resources []*SimpleAutoScalerResources `json:"resources"`

	// // last time of upscale
	// LastUpScaleTime metav1.Time `json:"lastUpScaleTime,omitempty"`

	// // last time of downscale
	// LastDownScaleTime metav1.Time `json:"lastDownScaleTime,omitempty"`
}

type SimpleAutoScalerCondition struct {
	// condition type
	Type SimpleAutoScalerConditionType `json:"type"`

	// condition status
	Status corev1.ConditionStatus `json:"status"`

	// LastTransitionTime is the last time the condition transitioned from on one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// message is a human-readable explaination containing details
	Message string `json:"message,omitempty"`
}

type SimpleAutoScalerConditionType string

const (
	// scaler match all resource and write it in status.resources
	MatchedAllResourceCondition SimpleAutoScalerConditionType = "MatchedAllResource"
	// checking policy filed and write it into currentFieldValue
	CheckPolicyFieldValueCondition SimpleAutoScalerConditionType = "CheckPolicyFieldValue"
	// checking policy scaleFactorObject and read the object values and write into scaleFactor
	GetScaleFactorCondition SimpleAutoScalerConditionType = "GetScaleFactor"
	// checking the desiredFieldValue whether in minAllowd and maxAllowd and write into status
	UpdatedFieldCondition SimpleAutoScalerConditionType = "UpdatedFieldCondition"
	// the scale whether can update the object compare the downscaleWindow and upscaleWindow, if okay
	// set lastUpScaleTime and lastDownScaleTime
	EndCooldownCondition SimpleAutoScalerConditionType = "EndCooldown"
	// when all ready it will update the target resource
	UpdateTargetResourceCondition SimpleAutoScalerConditionType = "UpdateTargetResource"
)

type SimpleAutoScalerResources struct {
	// Target points to the scale controls resources
	Target *autoscalingv1.CrossVersionObjectReference `json:"target,omitempty"`
	// record information of scale target
	ResourceFields []*ResourceStautsField `json:"resourceFields,omitempty"`
}

type ResourceStautsField struct {
	// name of field, be consistent in spec field
	Name string `json:"name,omitempty"`

	// path of fields
	Path string `json:"path,omitempty"`

	// from ScaleFactorObject read the value
	ScaleFactor string `json:"scaleFactor,omitempty"`

	// current value of field
	CurrentFieldValue resource.Quantity `json:"currentFieldValue,omitempty"`

	// desire value of field
	DesiredFieldValue resource.Quantity `json:"desiredFieldValue,omitempty"`

	// last time of upscale
	LastUpScaleTime metav1.Time `json:"lastUpScaleTime,omitempty"`

	// last time of downscale
	LastDownScaleTime metav1.Time `json:"lastDownScaleTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SimpleAutoScaler is the Schema for the simpleautoscalers API
type SimpleAutoScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimpleAutoScalerSpec   `json:"spec,omitempty"`
	Status SimpleAutoScalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SimpleAutoScalerList contains a list of SimpleAutoScaler
type SimpleAutoScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SimpleAutoScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SimpleAutoScaler{}, &SimpleAutoScalerList{})
}

// special function
func (s *SimpleAutoScalerStatus) AppendConditions(cond SimpleAutoScalerCondition) {
	var isExist bool
	for i, c := range s.Conditions {
		if c.Type == cond.Type {
			isExist = true
			if c.Status == cond.Status {
				cond.LastTransitionTime = metav1.NewTime(time.Now())
				s.Conditions[i] = cond
			}
			break
		}
	}
	if !isExist {
		cond.LastTransitionTime = metav1.NewTime(time.Now())
		s.Conditions = append(s.Conditions, cond)
	}
}

func (s *SimpleAutoScalerStatus) AddCondition(condtype SimpleAutoScalerConditionType, err error) {
	if err != nil {
		condition := SimpleAutoScalerCondition{
			Type:    condtype,
			Status:  "false",
			Message: err.Error(),
		}
		s.AppendConditions(condition)
	} else {
		condition := SimpleAutoScalerCondition{
			Type:   condtype,
			Status: "true",
		}
		s.AppendConditions(condition)
	}
}

// four condition status, true, false, unknow and nil string
func (s *SimpleAutoScalerStatus) GetConditionsStatus(cond SimpleAutoScalerConditionType) string {
	for _, c := range s.Conditions {
		if c.Type == cond {
			return string(c.Status)
		}
	}
	return ""
}
