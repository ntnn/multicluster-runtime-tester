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

// WhoamiSpec defines the desired state of Whoami.
type WhoamiSpec struct{}

// WhoamiStatus defines the observed state of Whoami.
type WhoamiStatus struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:format:URL
	// +required
	URL string `json:"url"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Whoami is the Schema for the whoamis API.
type Whoami struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WhoamiSpec   `json:"spec,omitempty"`
	Status WhoamiStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WhoamiList contains a list of Whoami.
type WhoamiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Whoami `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Whoami{}, &WhoamiList{})
}
