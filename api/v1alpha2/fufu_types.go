/*
Copyright 2022 huo.

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FufuSpec defines the desired state of Fufu
type FufuSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Fufu. Edit fufu_types.go to remove/update
	Color  string         `json:"color,required"`
	Weight string         `json:"weight,required"`
	Age    int            `json:"age,required"`
	Info   AdditionalInfo `json:"info,omitempty"`
}

type AdditionalInfo struct {
	Breed      string `json:"breed,omitempty"`
	Vaccinated bool   `json:"vaccinated,omitempty"`
}

// FufuStatus defines the observed state of Fufu
type FufuStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ExternalIP string `json:"externalIP,omitempty"`
	Replicas   int32  `json:"replicas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Color",type=string,JSONPath=`.spec.color`
//+kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.status.replicas`

// Fufu is the Schema for the fufus API
type Fufu struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FufuSpec   `json:"spec,omitempty"`
	Status FufuStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FufuList contains a list of Fufu
type FufuList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Fufu `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Fufu{}, &FufuList{})
}
