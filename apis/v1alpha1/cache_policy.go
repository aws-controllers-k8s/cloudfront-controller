// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

import (
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CachePolicySpec defines the desired state of CachePolicy.
//
// A cache policy.
//
// When it's attached to a cache behavior, the cache policy determines the following:
//
//   - The values that CloudFront includes in the cache key. These values can
//     include HTTP headers, cookies, and URL query strings. CloudFront uses
//     the cache key to find an object in its cache that it can return to the
//     viewer.
//
//   - The default, minimum, and maximum time to live (TTL) values that you
//     want objects to stay in the CloudFront cache.
//
// The headers, cookies, and query strings that are included in the cache key
// are also included in requests that CloudFront sends to the origin. CloudFront
// sends a request when it can't find a valid object in its cache that matches
// the request's cache key. If you want to send values to the origin but not
// include them in the cache key, use OriginRequestPolicy.
type CachePolicySpec struct {

	// A cache policy configuration.
	// +kubebuilder:validation:Required
	CachePolicyConfig *CachePolicyConfig `json:"cachePolicyConfig"`
}

// CachePolicyStatus defines the observed state of CachePolicy
type CachePolicyStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	// +kubebuilder:validation:Optional
	ACKResourceMetadata *ackv1alpha1.ResourceMetadata `json:"ackResourceMetadata"`
	// All CRs managed by ACK have a common `Status.Conditions` member that
	// contains a collection of `ackv1alpha1.Condition` objects that describe
	// the various terminal states of the CR and its backend AWS service API
	// resource
	// +kubebuilder:validation:Optional
	Conditions []*ackv1alpha1.Condition `json:"conditions"`
	// The current version of the cache policy.
	// +kubebuilder:validation:Optional
	ETag *string `json:"eTag,omitempty"`
	// The unique identifier for the cache policy.
	// +kubebuilder:validation:Optional
	ID *string `json:"id,omitempty"`
	// The date and time when the cache policy was last modified.
	// +kubebuilder:validation:Optional
	LastModifiedTime *metav1.Time `json:"lastModifiedTime,omitempty"`
}

// CachePolicy is the Schema for the CachePolicies API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type CachePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CachePolicySpec   `json:"spec,omitempty"`
	Status            CachePolicyStatus `json:"status,omitempty"`
}

// CachePolicyList contains a list of CachePolicy
// +kubebuilder:object:root=true
type CachePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CachePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CachePolicy{}, &CachePolicyList{})
}
