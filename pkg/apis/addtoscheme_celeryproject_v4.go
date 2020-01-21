package apis

import (
	v4 "github.com/RyanSiu1995/celery-operator/pkg/apis/celeryproject/v4"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v4.SchemeBuilder.AddToScheme)
}
