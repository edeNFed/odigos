/*
Copyright 2022.

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
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/odigos-io/odigos/api/odigos/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstrumentationLibraryStatusApplyConfiguration represents an declarative configuration of the InstrumentationLibraryStatus type for use
// with apply.
type InstrumentationLibraryStatusApplyConfiguration struct {
	Name                     *string                              `json:"name,omitempty"`
	Type                     *v1alpha1.InstrumentationLibraryType `json:"type,omitempty"`
	IdentifyingAttributes    []AttributeApplyConfiguration        `json:"identifyingAttributes,omitempty"`
	NonIdentifyingAttributes []AttributeApplyConfiguration        `json:"nonIdentifyingAttributes,omitempty"`
	Healthy                  *bool                                `json:"healthy,omitempty"`
	Message                  *string                              `json:"message,omitempty"`
	Reason                   *string                              `json:"reason,omitempty"`
	LastStatusTime           *v1.Time                             `json:"lastStatusTime,omitempty"`
}

// InstrumentationLibraryStatusApplyConfiguration constructs an declarative configuration of the InstrumentationLibraryStatus type for use with
// apply.
func InstrumentationLibraryStatus() *InstrumentationLibraryStatusApplyConfiguration {
	return &InstrumentationLibraryStatusApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *InstrumentationLibraryStatusApplyConfiguration) WithName(value string) *InstrumentationLibraryStatusApplyConfiguration {
	b.Name = &value
	return b
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *InstrumentationLibraryStatusApplyConfiguration) WithType(value v1alpha1.InstrumentationLibraryType) *InstrumentationLibraryStatusApplyConfiguration {
	b.Type = &value
	return b
}

// WithIdentifyingAttributes adds the given value to the IdentifyingAttributes field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the IdentifyingAttributes field.
func (b *InstrumentationLibraryStatusApplyConfiguration) WithIdentifyingAttributes(values ...*AttributeApplyConfiguration) *InstrumentationLibraryStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithIdentifyingAttributes")
		}
		b.IdentifyingAttributes = append(b.IdentifyingAttributes, *values[i])
	}
	return b
}

// WithNonIdentifyingAttributes adds the given value to the NonIdentifyingAttributes field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the NonIdentifyingAttributes field.
func (b *InstrumentationLibraryStatusApplyConfiguration) WithNonIdentifyingAttributes(values ...*AttributeApplyConfiguration) *InstrumentationLibraryStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithNonIdentifyingAttributes")
		}
		b.NonIdentifyingAttributes = append(b.NonIdentifyingAttributes, *values[i])
	}
	return b
}

// WithHealthy sets the Healthy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Healthy field is set to the value of the last call.
func (b *InstrumentationLibraryStatusApplyConfiguration) WithHealthy(value bool) *InstrumentationLibraryStatusApplyConfiguration {
	b.Healthy = &value
	return b
}

// WithMessage sets the Message field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Message field is set to the value of the last call.
func (b *InstrumentationLibraryStatusApplyConfiguration) WithMessage(value string) *InstrumentationLibraryStatusApplyConfiguration {
	b.Message = &value
	return b
}

// WithReason sets the Reason field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Reason field is set to the value of the last call.
func (b *InstrumentationLibraryStatusApplyConfiguration) WithReason(value string) *InstrumentationLibraryStatusApplyConfiguration {
	b.Reason = &value
	return b
}

// WithLastStatusTime sets the LastStatusTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastStatusTime field is set to the value of the last call.
func (b *InstrumentationLibraryStatusApplyConfiguration) WithLastStatusTime(value v1.Time) *InstrumentationLibraryStatusApplyConfiguration {
	b.LastStatusTime = &value
	return b
}
