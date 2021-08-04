// Copyright Contributors to the Open Cluster Management project

package v1alpha1

import (
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
	AuthRealmRef *corev1.LocalObjectReference `json:"authRealmRef,omitempty"`

	// Strategy to use for applying the AuthRealm to the managed clusters
	// +kubebuilder:validation:Enum=backplane;grc;hive
	// +required
	StrategyType StrategyType `json:"type"`

	//default to enforce but allow user to change, OPTIONAL - only used when type is grc
	// +kubebuilder:validation:Enum=enforce;inform
	RemediationType RemediationActionType `json:"remediationAction,omitempty"`

	// the list of TargetClusters for this strategy
	TargetClusters []TargetCluster `json:"targetClusters,omitempty"`
}

type StrategyType string

const (
	BackplaneStrategyType StrategyType = "backplane"
	GrcStrategyType       StrategyType = "grc"
	HiveStrategyType      StrategyType = "hive"
)

type RemediationActionType string

const (
	EnforceRemediationActionType RemediationActionType = "enforce"
	InformRemediationActionType  RemediationActionType = "inform"
)

// TargetCluster defined the cluster to apply the strategy to
type TargetCluster struct {
	// Name of the managed cluster
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// managed cluster ID for OAuth
	ClientID string `json:"clientID,omitempty"`

	// managed cluster secret for OAuth
	ClientSecretRef *corev1.LocalObjectReference `json:"clientSecret,omitempty"`

	// in case the user brings their own TLS cert for dex or rhsso
	CaCertificateRef *corev1.LocalObjectReference `json:"ca,omitempty"`
	Issuer           string                       `json:"issuer,omitempty"`
}

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
