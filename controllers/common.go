// Copyright Red Hat

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	identitatemv1alpha1 "github.com/identitatem/idp-client-api/api/identitatem/v1alpha1"

	identitatemdexv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"

	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
	workv1 "open-cluster-management.io/api/work/v1"

	"github.com/identitatem/idp-strategy-operator/pkg/helpers"
)

func (r *StrategyReconciler) addClientSecret(clusterName string, mw *workv1.ManifestWork) (*corev1.Secret, error) {
	//Build secret
	clientSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "idp-client-secret",
			Namespace: clusterName,
		},
		Data: map[string][]byte{
			"client-id":     []byte(clusterName),
			"client-secret": []byte(helpers.RandStringRunes(32)),
		},
	}

	clientSecretJSON, err := json.Marshal(clientSecret)
	if err != nil {
		return nil, err
	}

	mw.Spec.Workload.Manifests = append(mw.Spec.Workload.Manifests, workv1.Manifest{
		RawExtension: runtime.RawExtension{Raw: clientSecretJSON},
	})

	return clientSecret, nil

}

func (r *StrategyReconciler) addOAuth(dclusterName string, mw *workv1.ManifestWork) error {

	return nil
}

func (r *StrategyReconciler) createDexClient(authrealm *identitatemv1alpha1.AuthRealm, clusterName string, clientSecret *corev1.Secret) error {
	dexClientExists := true
	dexClient := &identitatemdexv1alpha1.DexClient{}
	if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: clusterName, Namespace: authrealm.Name}, dexClient); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		dexClientExists = false
		dexClient = &identitatemdexv1alpha1.DexClient{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: authrealm.Name,
			},
		}
	}

	dexClient.Spec.ClientID = string(clientSecret.Data["client-id"])
	dexClient.Spec.ClientSecret = string(clientSecret.Data["client-secret"])

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
	return nil
}

func inPlacementDecision(clusterName string, placementDecision *clusterv1alpha1.PlacementDecision) bool {
	for _, decision := range placementDecision.Status.Decisions {
		if decision.ClusterName == clusterName {
			return true
		}
	}
	return false
}
