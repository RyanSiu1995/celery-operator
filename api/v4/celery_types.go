/*
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

package v4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CelerySpec defines the desired state of Celery
type CelerySpec struct {
	Broker     CeleryBrokerSpec      `json:"broker,omitempty"`
	Workers    []CeleryWorkerSpec    `json:"workers,omitempty"`
	Schedulers []CelerySchedulerSpec `json:"schedulers,omitempty"`
	Image      string                `json:"image,omitempty"`
}

// CeleryStatus defines the observed state of Celery
type CeleryStatus struct {
	BrokerAddress string `json:"brokerAddress,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Celery is the Schema for the celeries API
type Celery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CelerySpec   `json:"spec,omitempty"`
	Status CeleryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CeleryList contains a list of Celery
type CeleryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Celery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Celery{}, &CeleryList{})
}
