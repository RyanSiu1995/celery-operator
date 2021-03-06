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

// CeleryBrokerSpec defines the desired state of CeleryBroker
type CeleryBrokerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CeleryBroker. Edit CeleryBroker_types.go to remove/update
	Type BrokerType `json:"type,omitempty"`
	// BrokerAddress defines the broker address for external broker type
	// If it is not `external` type, this item will be ignored
	BrokerAddress string `json:"brokerAddress,omitempty"`
}

// BrokerType defines the type of broker
type BrokerType string

const (
	// RedisBroker is to use a dynamic redis instead within cluster
	RedisBroker BrokerType = "redis"
	// ExternalBroker is to use an external broker with given string
	ExternalBroker BrokerType = "external"
)

// CeleryBrokerStatus defines the observed state of CeleryBroker
type CeleryBrokerStatus struct {
	BrokerAddress string `json:"brokerAddress,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CeleryBroker is the Schema for the celerybrokers API
type CeleryBroker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeleryBrokerSpec   `json:"spec,omitempty"`
	Status CeleryBrokerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CeleryBrokerList contains a list of CeleryBroker
type CeleryBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeleryBroker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CeleryBroker{}, &CeleryBrokerList{})
}
