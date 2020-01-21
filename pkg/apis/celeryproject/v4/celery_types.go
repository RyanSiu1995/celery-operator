package v4

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CelerySpec defines the desired state of Celery
type CelerySpec struct {
	Broker  CeleryBroker   `json:"broker,omitempty"`
	Workers []CeleryWorker `json:"workers,omitempty"`
}

// CeleryBroker defines the property of broker
type CeleryBroker struct {
	// Type defines the type of broker
	Type BrokerType `json:"type,omitempty"`
	// BrokerString defines the broker address for external broker type
	// If it is not `external` type, this item will be ignored
	BrokerString string `json:"brokerString,omitempty"`
}

// BrokerType defines the type of broker
type BrokerType string

// CeleryWorker defines the behavior of workers
type CeleryWorker struct {
	// DesiredNumber defines the number of worker if autoscaling is disabled
	DesiredNumber int `json:"desiredNumber,omitempty"`
	// Autoscaling defines the existence of HPA in celery worker
	Autoscaling bool `json:"autoscaling,omitempty"`
	// Min defines the minimum of workers if autoscaling is enabled
	Min int `json:"min,omitempty"`
	// Max defines the maximum of workers if autoscaling is enabled
	Max int `json:"max,omitempty"`
	// Target Queues defines the target queues these workers will handle
	TargetQueues []string `json:"targetQueues,omitempty"`
	// Resource defines the resources specification for these workers
	Resource corev1.ResourceRequirements `json:"resource,omitempty"`
}

const (
	// RedisBroker is to use a dynamic redis instead within cluster
	RedisBroker BrokerType = "redis"
	// ExternalBroker is to use an external broker with given string
	ExternalBroker BrokerType = "external"
)

// CeleryStatus defines the observed state of Celery
type CeleryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Celery is the Schema for the celeries API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=celeries,scope=Namespaced
type Celery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CelerySpec   `json:"spec,omitempty"`
	Status CeleryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CeleryList contains a list of Celery
type CeleryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Celery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Celery{}, &CeleryList{})
}
