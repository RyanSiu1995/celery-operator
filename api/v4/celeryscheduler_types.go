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

// CelerySchedulerSpec defines the desired state of CeleryScheduler
type CelerySchedulerSpec struct {
	Image string `json:"image,omitempty"`
	// SchedulerClass defines the target scheduler class to use
	SchedulerClass string `json:"schedulerClass,omitempty"`
	// AppName defines the target app instance to use
	AppName string `json:"appName,omitempty"`
	// DesiredNumber defines the number of worker if autoscaling is disabled
	Replicas int `json:"replicas,omitempty"`
	// Resources defines the resources specification for these workers
	Resources     corev1.ResourceRequirements `json:"resources,omitempty"`
	BrokerAddress string                      `json:"brokerAddress,omitempty"`
}

// CelerySchedulerStatus defines the observed state of CeleryScheduler
type CelerySchedulerStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CeleryScheduler is the Schema for the celeryschedulers API
type CeleryScheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CelerySchedulerSpec   `json:"spec,omitempty"`
	Status CelerySchedulerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CelerySchedulerList contains a list of CeleryScheduler
type CelerySchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeleryScheduler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CeleryScheduler{}, &CelerySchedulerList{})
}
