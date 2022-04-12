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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RecommendationScaleFactorSpec defines the desired state of RecommendationScaleFactor
type RecommendationScaleFactorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Reference of resources
	Ref *autoscalingv1.CrossVersionObjectReference `json:"ref,omitempty"`

	// expression of query
	Query string `json:"query,omitempty"`

	// desiredValue
	DesiredValue string `json:"desiredValue,omitempty"`

	// cooldown is the time of calcution, only affect statusFactor update and not affect the currentValue update
	Cooldown string `json:"cooldown,omitempty"`

	// minScope is the percentage of desireValue range
	MinScope int `json:"minScope,omitempty"`
}

// RecommendationScaleFactorStatus defines the observed state of RecommendationScaleFactor
type RecommendationScaleFactorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// result of query current value
	CurrentValue string `json:"currentValue,omitempty"`

	// Calculated by currentvalue and desiredvalue
	ScaleFactor string `json:"scaleFactor,omitempty"`

	// Conditions is set of conditions required for the recommendation scale
	Conditions []RecommendationScaleFactorCondition `json:"conditions,omitempty"`

	// last cooldown time
	LastCoolDownTime metav1.Time `json:"lastCoolDownTime,omitempty"`
}

type RecommendationScaleFactorCondition struct {
	// condition type
	Type RecommendationScaleFactorConditionType `json:"type"`

	// condition status
	Status corev1.ConditionStatus `json:"status"`

	// LastTransitionTime is the last time the condition transitioned from on one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// message is a human-readable explaination containing details
	Message string `json:"message,omitempty"`
}

type RecommendationScaleFactorConditionType string

const (
	QueryExistCondition             RecommendationScaleFactorConditionType = "QueryExist"
	ValidRecommendValueCondition    RecommendationScaleFactorConditionType = "ValidRecommendValue"
	RecommendationCoolDownCondition RecommendationScaleFactorConditionType = "Cooldown"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RecommendationScaleFactor is the Schema for the recommendationscalefactors API
type RecommendationScaleFactor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecommendationScaleFactorSpec   `json:"spec,omitempty"`
	Status RecommendationScaleFactorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RecommendationScaleFactorList contains a list of RecommendationScaleFactor
type RecommendationScaleFactorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RecommendationScaleFactor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RecommendationScaleFactor{}, &RecommendationScaleFactorList{})
}

// special function
func (s *RecommendationScaleFactorStatus) AppendConditions(cond RecommendationScaleFactorCondition) {
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
