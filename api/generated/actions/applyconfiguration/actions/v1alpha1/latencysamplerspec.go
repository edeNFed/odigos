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
	common "github.com/odigos-io/odigos/common"
)

// LatencySamplerSpecApplyConfiguration represents an declarative configuration of the LatencySamplerSpec type for use
// with apply.
type LatencySamplerSpecApplyConfiguration struct {
	ActionName       *string                             `json:"actionName,omitempty"`
	Notes            *string                             `json:"notes,omitempty"`
	Disabled         *bool                               `json:"disabled,omitempty"`
	Signals          []common.ObservabilitySignal        `json:"signals,omitempty"`
	EndpointsFilters []HttpRouteFilterApplyConfiguration `json:"endpoints_filters,omitempty"`
}

// LatencySamplerSpecApplyConfiguration constructs an declarative configuration of the LatencySamplerSpec type for use with
// apply.
func LatencySamplerSpec() *LatencySamplerSpecApplyConfiguration {
	return &LatencySamplerSpecApplyConfiguration{}
}

// WithActionName sets the ActionName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ActionName field is set to the value of the last call.
func (b *LatencySamplerSpecApplyConfiguration) WithActionName(value string) *LatencySamplerSpecApplyConfiguration {
	b.ActionName = &value
	return b
}

// WithNotes sets the Notes field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Notes field is set to the value of the last call.
func (b *LatencySamplerSpecApplyConfiguration) WithNotes(value string) *LatencySamplerSpecApplyConfiguration {
	b.Notes = &value
	return b
}

// WithDisabled sets the Disabled field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Disabled field is set to the value of the last call.
func (b *LatencySamplerSpecApplyConfiguration) WithDisabled(value bool) *LatencySamplerSpecApplyConfiguration {
	b.Disabled = &value
	return b
}

// WithSignals adds the given value to the Signals field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Signals field.
func (b *LatencySamplerSpecApplyConfiguration) WithSignals(values ...common.ObservabilitySignal) *LatencySamplerSpecApplyConfiguration {
	for i := range values {
		b.Signals = append(b.Signals, values[i])
	}
	return b
}

// WithEndpointsFilters adds the given value to the EndpointsFilters field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the EndpointsFilters field.
func (b *LatencySamplerSpecApplyConfiguration) WithEndpointsFilters(values ...*HttpRouteFilterApplyConfiguration) *LatencySamplerSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithEndpointsFilters")
		}
		b.EndpointsFilters = append(b.EndpointsFilters, *values[i])
	}
	return b
}
