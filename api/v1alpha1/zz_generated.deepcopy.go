// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/autoscaling/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecommendationScaleFactor) DeepCopyInto(out *RecommendationScaleFactor) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecommendationScaleFactor.
func (in *RecommendationScaleFactor) DeepCopy() *RecommendationScaleFactor {
	if in == nil {
		return nil
	}
	out := new(RecommendationScaleFactor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RecommendationScaleFactor) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecommendationScaleFactorCondition) DeepCopyInto(out *RecommendationScaleFactorCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecommendationScaleFactorCondition.
func (in *RecommendationScaleFactorCondition) DeepCopy() *RecommendationScaleFactorCondition {
	if in == nil {
		return nil
	}
	out := new(RecommendationScaleFactorCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecommendationScaleFactorList) DeepCopyInto(out *RecommendationScaleFactorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RecommendationScaleFactor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecommendationScaleFactorList.
func (in *RecommendationScaleFactorList) DeepCopy() *RecommendationScaleFactorList {
	if in == nil {
		return nil
	}
	out := new(RecommendationScaleFactorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RecommendationScaleFactorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecommendationScaleFactorSpec) DeepCopyInto(out *RecommendationScaleFactorSpec) {
	*out = *in
	if in.Ref != nil {
		in, out := &in.Ref, &out.Ref
		*out = new(v1.CrossVersionObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecommendationScaleFactorSpec.
func (in *RecommendationScaleFactorSpec) DeepCopy() *RecommendationScaleFactorSpec {
	if in == nil {
		return nil
	}
	out := new(RecommendationScaleFactorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecommendationScaleFactorStatus) DeepCopyInto(out *RecommendationScaleFactorStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]RecommendationScaleFactorCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.LastCoolDownTime.DeepCopyInto(&out.LastCoolDownTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecommendationScaleFactorStatus.
func (in *RecommendationScaleFactorStatus) DeepCopy() *RecommendationScaleFactorStatus {
	if in == nil {
		return nil
	}
	out := new(RecommendationScaleFactorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceScaleFactor) DeepCopyInto(out *ResourceScaleFactor) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceScaleFactor.
func (in *ResourceScaleFactor) DeepCopy() *ResourceScaleFactor {
	if in == nil {
		return nil
	}
	out := new(ResourceScaleFactor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceStautsField) DeepCopyInto(out *ResourceStautsField) {
	*out = *in
	out.CurrentFieldValue = in.CurrentFieldValue.DeepCopy()
	out.DesiredFieldValue = in.DesiredFieldValue.DeepCopy()
	in.LastUpScaleTime.DeepCopyInto(&out.LastUpScaleTime)
	in.LastDownScaleTime.DeepCopyInto(&out.LastDownScaleTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceStautsField.
func (in *ResourceStautsField) DeepCopy() *ResourceStautsField {
	if in == nil {
		return nil
	}
	out := new(ResourceStautsField)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScaleResourcePolicy) DeepCopyInto(out *ScaleResourcePolicy) {
	*out = *in
	if in.Update != nil {
		in, out := &in.Update, &out.Update
		*out = new(ScaleUpdateParameter)
		(*in).DeepCopyInto(*out)
	}
	if in.ScaleFactorObject != nil {
		in, out := &in.ScaleFactorObject, &out.ScaleFactorObject
		*out = new(ResourceScaleFactor)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScaleResourcePolicy.
func (in *ScaleResourcePolicy) DeepCopy() *ScaleResourcePolicy {
	if in == nil {
		return nil
	}
	out := new(ScaleResourcePolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScaleTargetResourceReference) DeepCopyInto(out *ScaleTargetResourceReference) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScaleTargetResourceReference.
func (in *ScaleTargetResourceReference) DeepCopy() *ScaleTargetResourceReference {
	if in == nil {
		return nil
	}
	out := new(ScaleTargetResourceReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScaleUpdateParameter) DeepCopyInto(out *ScaleUpdateParameter) {
	*out = *in
	out.MinAllowed = in.MinAllowed.DeepCopy()
	out.MaxAllowed = in.MaxAllowed.DeepCopy()
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScaleUpdateParameter.
func (in *ScaleUpdateParameter) DeepCopy() *ScaleUpdateParameter {
	if in == nil {
		return nil
	}
	out := new(ScaleUpdateParameter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleAutoScaler) DeepCopyInto(out *SimpleAutoScaler) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimpleAutoScaler.
func (in *SimpleAutoScaler) DeepCopy() *SimpleAutoScaler {
	if in == nil {
		return nil
	}
	out := new(SimpleAutoScaler)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SimpleAutoScaler) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleAutoScalerCondition) DeepCopyInto(out *SimpleAutoScalerCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimpleAutoScalerCondition.
func (in *SimpleAutoScalerCondition) DeepCopy() *SimpleAutoScalerCondition {
	if in == nil {
		return nil
	}
	out := new(SimpleAutoScalerCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleAutoScalerList) DeepCopyInto(out *SimpleAutoScalerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SimpleAutoScaler, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimpleAutoScalerList.
func (in *SimpleAutoScalerList) DeepCopy() *SimpleAutoScalerList {
	if in == nil {
		return nil
	}
	out := new(SimpleAutoScalerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SimpleAutoScalerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleAutoScalerResources) DeepCopyInto(out *SimpleAutoScalerResources) {
	*out = *in
	if in.Target != nil {
		in, out := &in.Target, &out.Target
		*out = new(v1.CrossVersionObjectReference)
		**out = **in
	}
	if in.ResourceFields != nil {
		in, out := &in.ResourceFields, &out.ResourceFields
		*out = make([]*ResourceStautsField, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ResourceStautsField)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimpleAutoScalerResources.
func (in *SimpleAutoScalerResources) DeepCopy() *SimpleAutoScalerResources {
	if in == nil {
		return nil
	}
	out := new(SimpleAutoScalerResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleAutoScalerSpec) DeepCopyInto(out *SimpleAutoScalerSpec) {
	*out = *in
	if in.TargetRef != nil {
		in, out := &in.TargetRef, &out.TargetRef
		*out = new(ScaleTargetResourceReference)
		(*in).DeepCopyInto(*out)
	}
	if in.Policy != nil {
		in, out := &in.Policy, &out.Policy
		*out = make([]*ScaleResourcePolicy, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ScaleResourcePolicy)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimpleAutoScalerSpec.
func (in *SimpleAutoScalerSpec) DeepCopy() *SimpleAutoScalerSpec {
	if in == nil {
		return nil
	}
	out := new(SimpleAutoScalerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleAutoScalerStatus) DeepCopyInto(out *SimpleAutoScalerStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]SimpleAutoScalerCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]*SimpleAutoScalerResources, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(SimpleAutoScalerResources)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimpleAutoScalerStatus.
func (in *SimpleAutoScalerStatus) DeepCopy() *SimpleAutoScalerStatus {
	if in == nil {
		return nil
	}
	out := new(SimpleAutoScalerStatus)
	in.DeepCopyInto(out)
	return out
}