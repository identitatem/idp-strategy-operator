// Copyright Red Hat

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	identitatemmgmtv1alpha1 "github.com/identitatem/idp-mgmt-operator/api/identitatem/v1alpha1"
	identitatemstrategyv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"

	clusterv1 "open-cluster-management.io/api/cluster/v1"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
	workv1 "open-cluster-management.io/api/work/v1"
)

const (
	backplaneManifestWorkName string = "idp-backplane"
)

func (r *StrategyReconciler) backplanePlacementStrategy(strategy *identitatemstrategyv1alpha1.Strategy,
	authrealm *identitatemmgmtv1alpha1.AuthRealm,
	placement *clusterv1alpha1.Placement,
	placementStrategy *clusterv1alpha1.Placement) error {
	// Append any additional predicates the AuthRealm already had on it's Placement
	placementStrategy.Spec.Predicates = placement.Spec.Predicates

	return nil
}

//DV
//backplaneStrategy generates resources for the Backplane strategy
func (r *StrategyReconciler) backplaneStrategy(strategy *identitatemstrategyv1alpha1.Strategy,
	authrealm *identitatemmgmtv1alpha1.AuthRealm,
	placement *clusterv1alpha1.Placement,
	placementDecision *clusterv1alpha1.PlacementDecision) error {
	//Get list of managedcluster
	mcs := &clusterv1.ManagedClusterList{}
	if err := r.Client.List(context.TODO(), mcs); err != nil {
		return err
	}
	//Loop on all managedcluster
	for _, mc := range mcs.Items {
		//Check if exists
		mw := &workv1.ManifestWork{}
		mwExists := true
		if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: backplaneManifestWorkName, Namespace: mc.Name}, mw); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			mwExists = false
			mw = &workv1.ManifestWork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      backplaneManifestWorkName,
					Namespace: mc.Name,
				},
			}
		}

		//If not in placementdecision then delete the manifestwork
		if mwExists && !inPlacementDecision(mc.Name, placementDecision) {
			if err := r.Client.Delete(context.TODO(), mw); err != nil {
				return err
			}
			break
		}

		//Create manifestwork
		mw.Spec.Workload.Manifests = make([]workv1.Manifest, 0)

		clientSecret, err := r.addClientSecret(mc.Name, mw)
		if err != nil {
			return err
		}

		if err := r.addOAuth(mc.Name, mw); err != nil {
			return err
		}

		switch mwExists {
		case true:
			if err := r.Client.Update(context.TODO(), mw); err != nil {
				return err
			}
		case false:
			if err := r.Client.Create(context.Background(), mw); err != nil {
				return err
			}
		}

		if err := r.createDexClient(authrealm, mc.Name, clientSecret); err != nil {
			return err
		}

	}
	return nil
}
