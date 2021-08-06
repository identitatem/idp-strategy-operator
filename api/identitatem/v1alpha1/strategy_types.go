// Copyright Contributors to the Open Cluster Management project

package v1alpha1

import (
	//policyv1 "github.com/open-cluster-management/governance-policy-propagator/pkg/apis/policy/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StrategySpec defines the desired state of Strategy
type StrategySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make generate-clients" to regenerate code after modifying this file

	// Authentication Realm
	// +required
	AuthRealm *corev1.LocalObjectReference `json:"authRealm,omitempty"`

	// Strategy to use for applying the AuthRealm to the managed clusters
	// +kubebuilder:validation:Enum=backplane;grc
	// +required
	StrategyType StrategyType `json:"type"`
}

type StrategyType string

const (
	BackplaneStrategyType StrategyType = "backplane"
	GrcStrategyType       StrategyType = "grc"
)

// StrategyStatus defines the observed state of Strategy
type StrategyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make generate-clients" to regenerate code after modifying this file

	// Conditions contains the different condition statuses for this AuthRealm.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

const (
	StrategyFailed  string = "Failed"
	StrategySucceed string = "Succeed"
)

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
