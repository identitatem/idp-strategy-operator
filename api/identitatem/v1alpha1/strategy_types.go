// Copyright Contributors to the Open Cluster Management project

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StrategySpec defines the desired state of Strategy
type StrategySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make generate-clients" to regenerate code after modifying this file

	// Foo is an example field of Strategy. Edit strategy_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// StrategyStatus defines the observed state of Strategy
type StrategyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make generate-clients" to regenerate code after modifying this file
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Strategy is the Schema for the strategies API
type Strategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StrategySpec   `json:"spec,omitempty"`
	Status StrategyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StrategyList contains a list of Strategy
type StrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Strategy `json:"items"`
}
