// Copyright Red Hat

package controllers

import (
	identitatemstrategyv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"

	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
)

//DV
//grcStrategy generates resources for the GRC strategy
func (r *StrategyReconciler) grcStrategy(strategy *identitatemstrategyv1alpha1.Strategy,
	placement *clusterv1alpha1.Placement,
	placementDecision *clusterv1alpha1.PlacementDecision) error {
	return nil
}
