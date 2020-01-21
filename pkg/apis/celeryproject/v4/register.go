// NOTE: Boilerplate only.  Ignore this file.

// Package v4 contains API Schema definitions for the celeryproject v4 API group
// +k8s:deepcopy-gen=package,register
// +groupName=celeryproject.org
package v4

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "celeryproject.org", Version: "v4"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)
