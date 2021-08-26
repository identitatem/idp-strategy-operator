// Copyright Red Hat

// +build functional

package functional

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	dexoperatorconfig "github.com/identitatem/dex-operator/config"
	identitatemclientset "github.com/identitatem/idp-client-api/api/client/clientset/versioned"
	identitatemv1alpha1 "github.com/identitatem/idp-client-api/api/identitatem/v1alpha1"
	idpconfig "github.com/identitatem/idp-client-api/config"
	openshiftconfigv1 "github.com/openshift/api/config/v1"
	clientsetcluster "open-cluster-management.io/api/client/cluster/clientset/versioned"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
	workv1 "open-cluster-management.io/api/work/v1"
	clusteradmapply "open-cluster-management.io/clusteradm/pkg/helpers/apply"
)

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)

}

var identitattemClientSet *identitatemclientset.Clientset
var clientSetCluster *clientsetcluster.Clientset
var k8sClient client.Client
var kubeClient *kubernetes.Clientset
var apiExtensionsClient *apiextensionsclient.Clientset
var dynamicClient dynamic.Interface

var cfg *rest.Config

var _ = Describe("Strategy", func() {
	AuthRealmName := "my-authrealm"
	AuthRealmNameSpace := "my-authrealmns"
	CertificatesSecretRef := "my-certs"
	StrategyName := AuthRealmName + "-backplane"
	// PlacementName := StrategyName
	ClusterName := "my-cluster"

	BeforeEach(func() {
		logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter)))
		SetDefaultEventuallyTimeout(20 * time.Second)
		SetDefaultEventuallyPollingInterval(1 * time.Second)

		kubeConfigFile := os.Getenv("KUBECONFIG")
		if len(kubeConfigFile) == 0 {
			home := homedir.HomeDir()
			kubeConfigFile = filepath.Join(home, ".kube", "config")
		}
		klog.Infof("KUBECONFIG=%s", kubeConfigFile)
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())

		kubeClient, err = kubernetes.NewForConfig(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeClient).ToNot(BeNil())
		apiExtensionsClient, err = apiextensionsclient.NewForConfig(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeClient).ToNot(BeNil())
		dynamicClient, err = dynamic.NewForConfig(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeClient).ToNot(BeNil())

		identitattemClientSet, err = identitatemclientset.NewForConfig(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(identitattemClientSet).ToNot(BeNil())

		clientSetCluster, err = clientsetcluster.NewForConfig(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(clientSetCluster).ToNot(BeNil())

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient).NotTo(BeNil())

		readerIDP := idpconfig.GetScenarioResourcesReader()
		applierBuilder := &clusteradmapply.ApplierBuilder{}
		applier := applierBuilder.WithClient(kubeClient, apiExtensionsClient, dynamicClient).Build()

		files := []string{
			"crd/bases/identityconfig.identitatem.io_authrealms.yaml",
		}
		_, err = applier.ApplyDirectly(readerIDP, nil, false, "", files...)
		Expect(err).Should(BeNil())

		readerDex := dexoperatorconfig.GetScenarioResourcesReader()
		files = []string{
			"crd/bases/auth.identitatem.io_dexclients.yaml",
		}
		_, err = applier.ApplyDirectly(readerDex, nil, false, "", files...)
		Expect(err).Should(BeNil())

		err = identitatemv1alpha1.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		err = clusterv1alpha1.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		err = clusterv1.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		err = workv1.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		err = openshiftconfigv1.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
	})

	It("process a Strategy", func() {
		By("creation test namespace", func() {
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
							Name:          "example.com",
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
			authRealm, err = identitattemClientSet.IdentityconfigV1alpha1().AuthRealms(AuthRealmNameSpace).Create(context.TODO(), authRealm, metav1.CreateOptions{})
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
		By("Create managedCluster", func() {
			mc := &clusterv1.ManagedCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: ClusterName,
				},
			}
			err := k8sClient.Create(context.TODO(), mc)
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
			_, err := identitattemClientSet.IdentityconfigV1alpha1().Strategies(AuthRealmNameSpace).Create(context.TODO(), strategy, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		Eventually(func() error {
			strategy, err := identitattemClientSet.IdentityconfigV1alpha1().Strategies(AuthRealmNameSpace).Get(context.TODO(), StrategyName, metav1.GetOptions{})
			if err != nil {
				logf.Log.Info("Error while reading strategy", "Error", err)
				return err
			}

			if len(strategy.Spec.Type) == 0 || strategy.Spec.Type != identitatemv1alpha1.BackplaneStrategyType {
				logf.Log.Info("Strategy Type is still wrong!")
				return fmt.Errorf("Strategy %s/%s not processed", strategy.Namespace, strategy.Name)
			}
			return nil
		}, 30, 1).Should(BeNil())

		By("Create Placement Decision CR", func() {
			placementDecision := &clusterv1alpha1.PlacementDecision{
				ObjectMeta: metav1.ObjectMeta{
					Name:      StrategyName,
					Namespace: AuthRealmNameSpace,
				},
			}
			placementDecision, err := clientSetCluster.ClusterV1alpha1().PlacementDecisions(AuthRealmNameSpace).
				Create(context.TODO(), placementDecision, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			placementDecision.Status.Decisions = []clusterv1alpha1.ClusterDecision{
				{
					ClusterName: ClusterName,
				},
			}
			_, err = clientSetCluster.ClusterV1alpha1().PlacementDecisions(AuthRealmNameSpace).
				UpdateStatus(context.TODO(), placementDecision, metav1.UpdateOptions{})
			Expect(err).To(BeNil())
		})

		Eventually(func() error {
			strategy, err := identitattemClientSet.IdentityconfigV1alpha1().Strategies(AuthRealmNameSpace).Get(context.TODO(), StrategyName, metav1.GetOptions{})
			if err != nil {
				logf.Log.Info("Error while reading strategy", "Error", err)
				return err
			}

			if len(strategy.Spec.Type) == 0 || strategy.Spec.Type != identitatemv1alpha1.BackplaneStrategyType {
				logf.Log.Info("Strategy Type is still wrong!")
				return fmt.Errorf("Strategy %s/%s not processed", strategy.Namespace, strategy.Name)
			}
			return nil
		}, 30, 1).Should(BeNil())

	})
})
