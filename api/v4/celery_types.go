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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CelerySpec defines the desired state of Celery
type CelerySpec struct {
	Broker  CeleryBroker   `json:"broker,omitempty"`
	Workers []CeleryWorker `json:"workers,omitempty"`
	Image   string         `json:"image,omitempty"`
}

// CeleryBroker defines the property of broker
type CeleryBroker struct {
	// Type defines the type of broker
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

// CeleryWorker defines the behavior of workers
type CeleryWorker struct {
	// DesiredNumber defines the number of worker if autoscaling is disabled
	Replicas int `json:"replicas,omitempty"`
	// Autoscaling defines the existence of HPA in celery worker
	Autoscaling bool `json:"autoscaling,omitempty"`
	// Min defines the minimum of workers if autoscaling is enabled
	Min int `json:"min,omitempty"`
	// Max defines the maximum of workers if autoscaling is enabled
	Max int `json:"max,omitempty"`
	// Target Queues defines the target queues these workers will handle
	TargetQueues []string `json:"targetQueues,omitempty"`
	// Resources defines the resources specification for these workers
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// AppName defines the target app instance to use
	AppName string `json:"appName,omitempty"`
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
