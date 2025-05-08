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

// APIServiceExportSpec defines the desired state of APIServiceExport.
type APIServiceExportSpec struct {
	// TODO Consumer instead of Target would be more apt
	Target           ExportTarget      `json:"target"`
	Resources        []ResourceRef     `json:"resources,omitempty"`
	PermissionClaims []PermissionClaim `json:"permissionClaims,omitempty"`
}

type ExportTarget struct {

	// TODO cluster is just the name of the cluster in
	// multicluster-runtime for now. This ofc does not work in e.g.
	// a multi-cluster multi-controller environment, where each
	// controller might have their own abstract name for a cluster.
	//
	// During the setup (when APIServiceExportRequest is created) the
	// creator should take care of any RBAC and service accounts
	// required and place kubeconfigs in the source and/or target
	// clusters for the controller(s) to use.
	// The APIServiceExportRequest could list the cluster names
	// accordingly.

	// TODO Rename to ClusterName

	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
}

type ResourceRef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PermissionClaim struct {
	Group            string             `json:"group"`
	Resource         string             `json:"resource"`
	Policy           Policy             `json:"policy"`
	ResourceSelector []ResourceSelector `json:"resourceSelector"`
}

type Policy struct {
	Provider PermissionClaimPolicy `json:"provider"`
	Consumer PermissionClaimPolicy `json:"consumer"`
}

type PermissionClaimPolicy struct {
	// TODO enums?
	Sync   string `json:"sync"`
	Delete string `json:"delete"`
}

type ResourceSelector struct {
	Group     string                      `json:"group"`
	Resource  string                      `json:"resource"`
	Policy    Policy                      `json:"policy"`
	Selectors []ResourceSelectorReference `json:"selectors"`
}

type ResourceSelectorReference struct {
	Resource string   `json:"resource"`
	Group    string   `json:"group"`
	Type     string   `json:"type"`
	JSONPath string   `json:"jsonpath"`
	Verbs    []string `json:"verbs"`
}

// APIServiceExportStatus defines the observed state of APIServiceExport.
type APIServiceExportStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// APIServiceExport is the Schema for the apiserviceexports API.
type APIServiceExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIServiceExportSpec   `json:"spec,omitempty"`
	Status APIServiceExportStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// APIServiceExportList contains a list of APIServiceExport.
type APIServiceExportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIServiceExport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIServiceExport{}, &APIServiceExportList{})
}
