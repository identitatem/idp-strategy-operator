// Copyright Red Hat

package strategy

import (
	identitatemv1alpha1 "github.com/identitatem/idp-client-api/api/identitatem/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
)

func (r *StrategyReconciler) backplanePlacementStrategy(strategy *identitatemv1alpha1.Strategy,
	authrealm *identitatemv1alpha1.AuthRealm,
	placement *clusterv1alpha1.Placement,
	placementStrategy *clusterv1alpha1.Placement) error {
	// Append any additional predicates the AuthRealm already had on it's Placement
	//placementStrategy.Spec.Predicates = placement.Spec.Predicates
	// If an addon is disabled, the feature label will be removed from the cluster. So you can select clusters where GRC is disabled with the placement below
	// placementStrategy.Spec.Predicates = []clusterv1alpha1.ClusterPredicate{
	// 	{
	// 		RequiredClusterSelector: clusterv1alpha1.ClusterSelector{
	// 			LabelSelector: metav1.LabelSelector{
	// 				MatchExpressions: []metav1.LabelSelectorRequirement{
	// 					{
	// 						Key:      "feature.open-cluster-management.io/addon-policy-controller",
	// 						Operator: metav1.LabelSelectorOpDoesNotExist,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// Append any additional predicates the AuthRealm already had on it's Placement
	//select clusters where GRC is not available (including addon disabled, unhealthy, unreachable)
	placementStrategy.Spec.Predicates = []clusterv1alpha1.ClusterPredicate{
		{
			RequiredClusterSelector: clusterv1alpha1.ClusterSelector{
				LabelSelector: metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "feature.open-cluster-management.io/addon-policy-controller",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"available"},
						},
					},
				},
			},
		},
	}

	placementStrategy.Spec.Predicates = append(placementStrategy.Spec.Predicates, placement.Spec.Predicates...)

	return nil
}
