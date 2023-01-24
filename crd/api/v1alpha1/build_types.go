/*
Copyright 2023.

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

type Target struct {
	Source     string `json:"source_url,omitempty"`
	Commit     string `json:"commit_id,omitempty"`
	Dockerfile string `json:"dockerfile,omitempty"`
}

type Destination struct {
	Host  string `json:"address"`
	Image string `json:"image"`
	Tag   string `json:"tag"`
}

// BuildSpec defines the desired state of Build
type BuildSpec struct {
	Target      `json:"target,omitempty"`
	Destination `json:"destination,omitempty"`
}

// BuildStatus defines the observed state of Build
type BuildStatus struct {
	Success bool   `json:"success,omitempty"`
	Image   string `json:"image,omitempty"`
	Tag     string `json:"tag,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Build is the Schema for the builds API
type Build struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BuildSpec   `json:"spec,omitempty"`
	Status BuildStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BuildList contains a list of Build
type BuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Build `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Build{}, &BuildList{})
}
