package models

import "fmt"

// AppID uniquely identifies an application by its Kubernetes namespace, kind, and name.
type AppID struct {
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"` // deployment, daemonset, statefulset
	Name      string `json:"name"`
}

// String returns a string representation of the AppID.
func (a AppID) String() string {
	return fmt.Sprintf("%s/%s/%s", a.Namespace, a.Kind, a.Name)
}

// IsEmpty returns true if the AppID has no values set.
func (a AppID) IsEmpty() bool {
	return a.Namespace == "" && a.Kind == "" && a.Name == ""
}
