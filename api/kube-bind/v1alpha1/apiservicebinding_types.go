/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APIServiceBindingSpec defines the desired state of APIServiceBinding.
type APIServiceBindingSpec struct {
	// KubeconfigSecretRef SecretRef                `json:"kubeconfigSecretRef"`
	// PermissionClaims []BindingPermissionClaim `json:"permissionClaims"`
}

type SecretRef struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type BindingPermissionClaim struct {
	PermissionClaim `json:",inline"`
	Status          AcceptanceStatus `json:"status"`
}

type AcceptanceStatus string

const (
	AcceptancePending AcceptanceStatus = "Pending"
	Declined          AcceptanceStatus = "Declined"
	Valid             AcceptanceStatus = "Valid"
	Accepted          AcceptanceStatus = "Accepted"
)

// APIServiceBindingStatus defines the observed state of APIServiceBinding.
type APIServiceBindingStatus struct {
	BoundResources   []BindingResourceRef     `json:"boundResources"`
	PermissionClaims []BindingPermissionClaim `json:"permissionClaims"`
	Conditions       []metav1.Condition       `json:"conditions"`
}

type BindingResourceRef struct {
	Group    string `json:"group"`
	Resource string `json:"resource"`
}

type APIServiceBindingStatusType string

const (
	BindingPending  APIServiceBindingStatusType = "Pending"
	SecretValid     APIServiceBindingStatusType = "SecretValid"
	InformersSynced APIServiceBindingStatusType = "InformersSynced"
	Heartbeating    APIServiceBindingStatusType = "Heartbeating"
	Connected       APIServiceBindingStatusType = "Connected"
	Ready           APIServiceBindingStatusType = "Ready"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// APIServiceBinding is the Schema for the apiservicebindings API.
type APIServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIServiceBindingSpec   `json:"spec,omitempty"`
	Status APIServiceBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// APIServiceBindingList contains a list of APIServiceBinding.
type APIServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIServiceBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIServiceBinding{}, &APIServiceBindingList{})
}
