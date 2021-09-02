// Copyright Red Hat

// +build functional

package functional

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	dexv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"
	identitatemv1alpha1 "github.com/identitatem/idp-client-api/api/identitatem/v1alpha1"
	openshiftconfigv1 "github.com/openshift/api/config/v1"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
)

var _ = Describe("Strategy", func() {
	AuthRealmName := "my-authrealm"
	AuthRealmNameSpace := "my-authrealmns"
	CertificatesSecretRef := "my-certs"
	StrategyName := AuthRealmName + "-backplane"
	PlacementStrategyName := StrategyName
	ClusterName := "my-cluster"
	MyIDPName := "my-idp"

	It("process a Strategy", func() {
		By(fmt.Sprintf("creation of User namespace %s", AuthRealmNameSpace), func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: AuthRealmNameSpace,
				},
			}
			err := k8sClient.Create(context.TODO(), ns)
			Expect(err).To(BeNil())
		})
		var placement *clusterv1alpha1.Placement
		By("Creating placement", func() {
			placement = &clusterv1alpha1.Placement{
				ObjectMeta: metav1.ObjectMeta{
					Name:      AuthRealmName,
					Namespace: AuthRealmNameSpace,
				},
				Spec: clusterv1alpha1.PlacementSpec{
					Predicates: []clusterv1alpha1.ClusterPredicate{
						{
							RequiredClusterSelector: clusterv1alpha1.ClusterSelector{
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{
										"mylabel": "test",
									},
								},
							},
						},
					},
				},
			}
			var err error
			placement, err = clientSetCluster.ClusterV1alpha1().Placements(AuthRealmNameSpace).
				Create(context.TODO(), placement, metav1.CreateOptions{})
			Expect(err).To(BeNil())

		})
		var authRealm *identitatemv1alpha1.AuthRealm
		By("creating a AuthRealm CR", func() {
			var err error
			authRealm = &identitatemv1alpha1.AuthRealm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      AuthRealmName,
					Namespace: AuthRealmNameSpace,
				},
				Spec: identitatemv1alpha1.AuthRealmSpec{
					Type: identitatemv1alpha1.AuthProxyDex,
					CertificatesSecretRef: corev1.LocalObjectReference{
						Name: CertificatesSecretRef,
					},
					IdentityProviders: []openshiftconfigv1.IdentityProvider{
						{
							Name:          MyIDPName,
							MappingMethod: openshiftconfigv1.MappingMethodClaim,
							IdentityProviderConfig: openshiftconfigv1.IdentityProviderConfig{
								Type: openshiftconfigv1.IdentityProviderTypeGitHub,
								GitHub: &openshiftconfigv1.GitHubIdentityProvider{
									ClientID: "me",
								},
							},
						},
					},
					PlacementRef: corev1.LocalObjectReference{
						Name: placement.Name,
					},
				},
			}
			//DV reassign  to authRealm to get the extra info that kube set (ie:uuid as needed to set ownerref)
			authRealm, err = identitatemClientSet.IdentityconfigV1alpha1().AuthRealms(AuthRealmNameSpace).Create(context.TODO(), authRealm, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		By("Create a Backplane Strategy", func() {
			strategy := &identitatemv1alpha1.Strategy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      StrategyName,
					Namespace: AuthRealmNameSpace,
				},
				Spec: identitatemv1alpha1.StrategySpec{
					Type: identitatemv1alpha1.BackplaneStrategyType,
				},
			}
			controllerutil.SetOwnerReference(authRealm, strategy, scheme.Scheme)
			_, err := identitatemClientSet.IdentityconfigV1alpha1().Strategies(AuthRealmNameSpace).Create(context.TODO(), strategy, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		By("creation cluster namespace", func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: ClusterName,
				},
			}
			err := k8sClient.Create(context.TODO(), ns)
			Expect(err).To(BeNil())
		})
		By(fmt.Sprintf("Checking creation of strategy placement %s", PlacementStrategyName), func() {
			Eventually(func() error {
				_, err := clientSetCluster.ClusterV1alpha1().Placements(AuthRealmNameSpace).Get(context.TODO(), PlacementStrategyName, metav1.GetOptions{})
				if err != nil {
					if !errors.IsNotFound(err) {
						return err
					}
					logf.Log.Info("Placement", "Name", PlacementStrategyName, "Namespace", AuthRealmNameSpace)
					return err
				}
				return nil
			}, 30, 1).Should(BeNil())
			By("Checking strategy", func() {
				var err error
				strategy, err := identitatemClientSet.IdentityconfigV1alpha1().Strategies(AuthRealmNameSpace).Get(context.TODO(), StrategyName, metav1.GetOptions{})
				Expect(err).To(BeNil())
				Expect(strategy.Spec.PlacementRef.Name).Should(Equal(PlacementStrategyName))
			})
			By("Checking placement strategy", func() {
				_, err := clientSetCluster.ClusterV1alpha1().Placements(AuthRealmNameSpace).
					Get(context.TODO(), PlacementStrategyName, metav1.GetOptions{})
				Expect(err).To(BeNil())
				Expect(len(placement.Spec.Predicates)).Should(Equal(1))
			})
		})
	})
	It("process a PlacementDecision", func() {
		var placementStrategy *clusterv1alpha1.Placement
		By("Checking placement strategy", func() {
			var err error
			placementStrategy, err = clientSetCluster.ClusterV1alpha1().Placements(AuthRealmNameSpace).
				Get(context.TODO(), PlacementStrategyName, metav1.GetOptions{})
			Expect(err).To(BeNil())
		})
		By(fmt.Sprintf("creation of Dex namespace %s", AuthRealmName), func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: AuthRealmName,
				},
			}
			err := k8sClient.Create(context.TODO(), ns)
			Expect(err).To(BeNil())
		})

		var placementDecision *clusterv1alpha1.PlacementDecision
		By("Create Placement Decision CR", func() {
			placementDecision = &clusterv1alpha1.PlacementDecision{
				ObjectMeta: metav1.ObjectMeta{
					Name:      StrategyName,
					Namespace: AuthRealmNameSpace,
					Labels: map[string]string{
						clusterv1alpha1.PlacementLabel: placementStrategy.Name,
					},
				},
			}
			controllerutil.SetOwnerReference(placementStrategy, placementDecision, scheme.Scheme)
			var err error
			placementDecision, err = clientSetCluster.ClusterV1alpha1().PlacementDecisions(AuthRealmNameSpace).
				Create(context.TODO(), placementDecision, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			Eventually(func() error {
				placementDecision, err = clientSetCluster.ClusterV1alpha1().PlacementDecisions(AuthRealmNameSpace).Get(context.TODO(), StrategyName, metav1.GetOptions{})
				Expect(err).To(BeNil())

				placementDecision.Status.Decisions = []clusterv1alpha1.ClusterDecision{
					{
						ClusterName: ClusterName,
					},
				}
				_, err = clientSetCluster.ClusterV1alpha1().PlacementDecisions(AuthRealmNameSpace).
					UpdateStatus(context.TODO(), placementDecision, metav1.UpdateOptions{})
				return err
			}, 30, 1).Should(BeNil())
		})

		dexClientName := fmt.Sprintf("%s-%s", ClusterName, MyIDPName)
		By(fmt.Sprintf("Checking client secret %s", MyIDPName), func() {
			Eventually(func() error {
				clientSecret := &corev1.Secret{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: MyIDPName, Namespace: ClusterName}, clientSecret)
				if err != nil {
					if !errors.IsNotFound(err) {
						return err
					}
					logf.Log.Info("ClientSecret", "Name", MyIDPName, "Namespace", ClusterName)
					return err
				}
				return nil
			}, 30, 1).Should(BeNil())
		})
		By(fmt.Sprintf("Checking DexClient %s", dexClientName), func() {
			Eventually(func() error {
				dexClient := &dexv1alpha1.DexClient{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: dexClientName, Namespace: AuthRealmName}, dexClient)
				if err != nil {
					if !errors.IsNotFound(err) {
						return err
					}
					logf.Log.Info("DexClient", "Name", dexClientName, "Namespace", AuthRealmName)
					return err
				}
				return nil
			}, 30, 1).Should(BeNil())
		})
		By("Deleting the placementdecision", func() {
			err := clientSetCluster.ClusterV1alpha1().PlacementDecisions(AuthRealmNameSpace).Delete(context.TODO(), StrategyName, metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		})
		By(fmt.Sprintf("Checking client secret deletion %s", MyIDPName), func() {
			Eventually(func() error {
				clientSecret := &corev1.Secret{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: MyIDPName, Namespace: ClusterName}, clientSecret)
				if err != nil {
					if !errors.IsNotFound(err) {
						return err
					}
					return nil
				}
				return fmt.Errorf("clientSecret %s still exist", MyIDPName)
			}, 30, 1).Should(BeNil())
		})
		By(fmt.Sprintf("Checking DexClient deletion %s", dexClientName), func() {
			Eventually(func() error {
				dexClient := &dexv1alpha1.DexClient{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: dexClientName, Namespace: AuthRealmName}, dexClient)
				if err != nil {
					if !errors.IsNotFound(err) {
						return err
					}
					return nil
				}
				return fmt.Errorf("DexClient %s still exist", dexClientName)

			}, 30, 1).Should(BeNil())
		})
		By(fmt.Sprintf("Checking PlacementDecision deletion %s", dexClientName), func() {
			Eventually(func() error {
				_, err := clientSetCluster.ClusterV1alpha1().PlacementDecisions(AuthRealmNameSpace).Get(context.TODO(), StrategyName, metav1.GetOptions{})
				if err != nil {
					if !errors.IsNotFound(err) {
						return err
					}
					return nil
				}
				return fmt.Errorf("PlacementDecision %s still exist", StrategyName)

			}, 30, 1).Should(BeNil())
		})
		// By("Deleting the AuthRealm", func() {
		// 	err := identitatemClientSet.IdentityconfigV1alpha1().AuthRealms(AuthRealmNameSpace).Delete(context.TODO(), AuthRealmName, metav1.DeleteOptions{})
		// 	Expect(err).To(BeNil())
		// })
	})
})
