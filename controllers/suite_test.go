// Copyright Red Hat

package controllers

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	clientsetcluster "open-cluster-management.io/api/client/cluster/clientset/versioned"
	clientsetwork "open-cluster-management.io/api/client/work/clientset/versioned"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
	workv1 "open-cluster-management.io/api/work/v1"

	clientsetmgmt "github.com/identitatem/idp-mgmt-operator/api/client/clientset/versioned"
	identitatemmgmtv1alpha1 "github.com/identitatem/idp-mgmt-operator/api/identitatem/v1alpha1"

	clientsetstrategy "github.com/identitatem/idp-strategy-operator/api/client/clientset/versioned"
	identitatemstrategyv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"
	openshiftconfigv1 "github.com/openshift/api/config/v1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var clientSetMgmt *clientsetmgmt.Clientset
var clientSetStrategy *clientsetstrategy.Clientset
var clientSetCluster *clientsetcluster.Clientset
var clientSetWork *clientsetwork.Clientset
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases"),
			//DV added this line and copyed the authrealms CRD
			filepath.Join("..", "config", "crd", "external")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = identitatemstrategyv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = identitatemmgmtv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clusterv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = workv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	clientSetMgmt, err = clientsetmgmt.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(clientSetMgmt).ToNot(BeNil())

	clientSetStrategy, err = clientsetstrategy.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(clientSetStrategy).ToNot(BeNil())

	clientSetCluster, err = clientsetcluster.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(clientSetCluster).ToNot(BeNil())

	clientSetWork, err = clientsetwork.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(clientSetWork).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Process Strategy backplane: ", func() {
	AuthRealmName := "test-authrealm"
	AuthRealmNameSpace := "test"
	CertificatesSecretRef := "test-certs"
	StrategyName := "test-strategy"
	PlacementName := "test-placement"
	ClusterName := "mycluster"

	It("process a Strategy backplane CR", func() {
		By("creation test namespace", func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: AuthRealmNameSpace,
				},
			}
			err := k8sClient.Create(context.TODO(), ns)
			Expect(err).To(BeNil())
		})
		var authRealm = &identitatemmgmtv1alpha1.AuthRealm{}
		By("creating a AuthRealm CR", func() {
			//first create AuthRealm
			authRealm = &identitatemmgmtv1alpha1.AuthRealm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      AuthRealmName,
					Namespace: AuthRealmNameSpace,
				},
				Spec: identitatemmgmtv1alpha1.AuthRealmSpec{
					Type: identitatemmgmtv1alpha1.AuthProxyDex,
					CertificatesSecretRef: corev1.LocalObjectReference{
						Name: CertificatesSecretRef,
					},
					IdentityProviders: []identitatemmgmtv1alpha1.IdentityProvider{
						{
							GitHub: &openshiftconfigv1.GitHubIdentityProvider{},
						},
					},
					//DV add the placement specification as it will be used to create the final placement
					Placement: &identitatemmgmtv1alpha1.Placement{
						Name: PlacementName,
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
					},
				},
			}
			//DV reassign  to authRealm to get the extra info that kube set (ie:uuid as needed to set ownerref)
			var err error
			authRealm, err = clientSetMgmt.IdentityconfigV1alpha1().AuthRealms(AuthRealmNameSpace).Create(context.TODO(), authRealm, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		By("creating a Strategy CR", func() {
			strategy := &identitatemstrategyv1alpha1.Strategy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      StrategyName,
					Namespace: AuthRealmNameSpace,
					//DV Commented
					// OwnerReferences: []metav1.OwnerReference{
					// 	{
					// 		Name: AuthRealmName,
					// 	},
					// },
				},
				Spec: identitatemstrategyv1alpha1.StrategySpec{
					Type: identitatemstrategyv1alpha1.BackplaneStrategyType,
				},
			}

			//_, err := identitatemClientSet.IdentityconfigV1alpha1().Strategies("default").Create(context.TODO(), &strategy, metav1.CreateOptions{})
			//_, err := ClientSetStrategy.IdentityconfigV1alpha1().Strategies("default").Create(context.TODO(), &strategy, metav1.CreateOptions{})
			//_, err := ClientSetStrategy.IdentityconfigV1alpha1().Strategies("default").Create(context.TODO(), &strategy, metav1.CreateOptions{})

			//DV Added this to set the ownerref
			controllerutil.SetOwnerReference(authRealm, strategy, scheme.Scheme)

			tmp, err := clientSetStrategy.IdentityconfigV1alpha1().Strategies(AuthRealmNameSpace).Create(context.TODO(), strategy, metav1.CreateOptions{})
			if tmp == nil {
				//just put this here to get no complaints...but need to remove.  _ instead of tmp did not work above
			}

			Expect(err).To(BeNil())
		})
		//DV replace Eventually by By as no need for waiting as it is a method call.... my bad.
		By("Calling reconcile", func() {
			r := StrategyReconciler{
				Client: k8sClient,
				Log:    logf.Log,
				Scheme: scheme.Scheme,
			}

			req := ctrl.Request{}
			req.Name = StrategyName
			req.Namespace = AuthRealmNameSpace
			result, err := r.Reconcile(context.TODO(), req)
			Expect(err).To(BeNil())
			//DV Should be requeued as we wait for the PlacementDecision
			Expect(result.Requeue).To(BeTrue())
			Expect(result.RequeueAfter).To(Equal(time.Second * 10))
		})
		var strategy *identitatemstrategyv1alpha1.Strategy
		By("Checking strategy", func() {
			var err error
			strategy, err = clientSetStrategy.IdentityconfigV1alpha1().Strategies(AuthRealmNameSpace).Get(context.TODO(), StrategyName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			//DV No need as By now
			// if err != nil {
			// 	logf.Log.Info("Error while reading authrealm", "Error", err)
			// 	return err
			// }
			Expect(strategy.Spec.PlacementRef.Name).Should(Equal(PlacementName))
		})
		//DV Add check on placement
		By("Checking placement", func() {
			placement, err := clientSetCluster.ClusterV1alpha1().Placements(AuthRealmNameSpace).
				Get(context.TODO(), strategy.Spec.PlacementRef.Name, metav1.GetOptions{})
			Expect(err).To(BeNil())
			//DV No need as By now
			// if err != nil {
			// 	logf.Log.Info("Error while reading authrealm", "Error", err)
			// 	return err
			// }
			Expect(len(placement.Spec.Predicates)).Should(Equal(2))
		})
		By("Create Placement Decision CR", func() {
			placementDecision := &clusterv1alpha1.PlacementDecision{
				ObjectMeta: metav1.ObjectMeta{
					Name:      strategy.Spec.PlacementRef.Name,
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
		By("creation cluster namespace", func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: ClusterName,
				},
			}
			err := k8sClient.Create(context.TODO(), ns)
			Expect(err).To(BeNil())
		})
		By("Calling reconcile", func() {
			r := StrategyReconciler{
				Client: k8sClient,
				Log:    logf.Log,
				Scheme: scheme.Scheme,
			}

			req := ctrl.Request{}
			req.Name = StrategyName
			req.Namespace = AuthRealmNameSpace
			_, err := r.Reconcile(context.TODO(), req)
			Expect(err).To(BeNil())
		})
		By("Checking manifestwork", func() {
			mw, err := clientSetWork.WorkV1().ManifestWorks(ClusterName).Get(context.TODO(), "idp", metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(len(mw.Spec.Workload.Manifests)).To(Equal(1))
		})
	})
})
