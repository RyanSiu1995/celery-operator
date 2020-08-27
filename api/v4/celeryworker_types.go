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

// CeleryWorkerSpec defines the desired state of CeleryWorker
type CeleryWorkerSpec struct {
	// DesiredNumber defines the number of worker if autoscaling is disabled
	Replicas int `json:"replicas,omitempty"`
	// Target Queues defines the target queues these workers will handle
	TargetQueues []string `json:"targetQueues,omitempty"`
	// Resources defines the resources specification for these workers
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// AppName defines the target app instance to use
	AppName       string `json:"appName,omitempty"`
	BrokerAddress string `json:"brokerAddress,omitempty"`
	Image         string `json:"image,omitempty"`
}

// CeleryWorkerStatus defines the observed state of CeleryWorker
type CeleryWorkerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CeleryWorker is the Schema for the celeryworkers API
type CeleryWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeleryWorkerSpec   `json:"spec,omitempty"`
	Status CeleryWorkerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CeleryWorkerList contains a list of CeleryWorker
type CeleryWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeleryWorker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CeleryWorker{}, &CeleryWorkerList{})
}
