// Copyright Red Hat

package placementdecision

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	identitatemv1alpha1 "github.com/identitatem/idp-client-api/api/identitatem/v1alpha1"

	identitatemdexv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"

	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
	workv1 "open-cluster-management.io/api/work/v1"

	"github.com/identitatem/idp-strategy-operator/pkg/helpers"
)

// DO NOT REGEMERATE SECRET. READ THE ONE IN DEXCLIENT
// func (r *PlacementDecisionReconciler) addClientSecret(clusterName string, mw *workv1.ManifestWork) (*corev1.Secret, error) {
// 	//Build secret
// 	clientSecret := &corev1.Secret{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Secret",
// 			APIVersion: "v1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "idp-client-secret",
// 			Namespace: clusterName,
// 		},
// 		Data: map[string][]byte{
// 			"client-id":     []byte(clusterName),
// 			"client-secret": []byte(helpers.RandStringRunes(32)),
// 		},
// 	}

// 	clientSecretJSON, err := json.Marshal(clientSecret)
// 	if err != nil {
// 		return nil, err
// 	}

// 	mw.Spec.Workload.Manifests = append(mw.Spec.Workload.Manifests, workv1.Manifest{
// 		RawExtension: runtime.RawExtension{Raw: clientSecretJSON},
// 	})

// 	return clientSecret, nil

// }

func (r *PlacementDecisionReconciler) addOAuth(dclusterName string, mw *workv1.ManifestWork) error {

	return nil
}

func (r *PlacementDecisionReconciler) syncDexClients(authrealm *identitatemv1alpha1.AuthRealm, placementDecision *clusterv1alpha1.PlacementDecision) error {

	dexClients := &identitatemdexv1alpha1.DexClientList{}
	if err := r.Client.List(context.TODO(), dexClients, &client.ListOptions{Namespace: authrealm.Name}); err != nil {
		return err
	}
	for i, dexClient := range dexClients.Items {
		for _, idp := range authrealm.Spec.IdentityProviders {
			if !inPlacementDecision(dexClient.GetLabels()["cluster"], placementDecision) &&
				dexClient.GetLabels()["idp"] == idp.Name {
				if err := r.Client.Delete(context.TODO(), &dexClients.Items[i]); err != nil {
					return err
				}
			}
		}
	}
	for _, decision := range placementDecision.Status.Decisions {
		for _, idp := range authrealm.Spec.IdentityProviders {
			clusterName := decision.ClusterName
			dexClientExists := true
			dexClient := &identitatemdexv1alpha1.DexClient{}
			if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: clusterName, Namespace: authrealm.Name}, dexClient); err != nil {
				if !errors.IsNotFound(err) {
					return err
				}
				dexClientExists = false
				dexClient = &identitatemdexv1alpha1.DexClient{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%s", clusterName, idp.Name),
						Namespace: authrealm.Name,
						Labels: map[string]string{
							"cluster": clusterName,
							"idp":     idp.Name,
						},
					},
				}
			}

			dexClient.Spec.ClientID = string([]byte(clusterName))
			dexClient.Spec.ClientSecret = string([]byte(helpers.RandStringRunes(32)))

			apiServerURL, err := helpers.GetKubeAPIServerAddress(r.Client)
			if err != nil {
				return err
			}
			u, err := url.Parse(apiServerURL)
			if err != nil {
				return err
			}

			host, _, err := net.SplitHostPort(u.Host)
			if err != nil {
				return err
			}

			host = strings.Replace(host, "api", "apps", 1)

			redirectURI := fmt.Sprintf("%s://%s/oauth2callback/idpserver", u.Scheme, host)
			dexClient.Spec.RedirectURIs = []string{redirectURI}
			switch dexClientExists {
			case true:
				return r.Client.Update(context.TODO(), dexClient)
			case false:
				return r.Client.Create(context.Background(), dexClient)
			}
		}
	}
	return nil
}

func GetStrategyFromPlacementDecision(c client.Client, placementDecisionName, placementDecisionNamespace string) (*identitatemv1alpha1.Strategy, error) {
	//Placement decisions has the same name then PlacementDecision
	//and so we can search the placement from placement decsion
	return GetStrategyFromPlacement(c, placementDecisionName, placementDecisionNamespace)
}

func GetStrategyFromPlacement(c client.Client, placementName, placementNamespace string) (*identitatemv1alpha1.Strategy, error) {
	strategies := &identitatemv1alpha1.StrategyList{}
	if err := c.List(context.TODO(), strategies, &client.ListOptions{Namespace: placementNamespace}); err != nil {
		return nil, err
	}
	for _, strategy := range strategies.Items {
		if strategy.Spec.PlacementRef.Name == placementName {
			return &strategy, nil
		}
	}
	return nil, errors.NewNotFound(identitatemv1alpha1.Resource("strategies"), placementName)
	// strategy := &identitatemv1alpha1.Strategy{}
	// var ownerRef metav1.OwnerReference
	// //DV not needed
	// // placementInfo := &identitatemv1alpha1.Placement{}

	// //for _, or := range ownerRefs {
	// for _, or := range strategy.GetOwnerReferences() {

	// 	//TODO find a better way
	// 	if or.Kind == "Strategy" {
	// 		ownerRef = or
	// 		break
	// 	}
	// }
	// if err := c.Get(context.TODO(), client.ObjectKey{Name: ownerRef.Name, Namespace: strategy.Namespace}, strategy); err != nil {
	// 	return nil, err
	// }
	// return strategy, nil
}

func inPlacementDecision(clusterName string, placementDecision *clusterv1alpha1.PlacementDecision) bool {
	for _, decision := range placementDecision.Status.Decisions {
		if decision.ClusterName == clusterName {
			return true
		}
	}
	return false
}
