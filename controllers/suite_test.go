// Copyright Red Hat

package controllers

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	strategyclientset "github.com/identitatem/idp-strategy-operator/api/client/clientset/versioned"
	identitatemv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var identitatemClientSet *strategyclientset.Clientset
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
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = identitatemv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	identitatemClientSet, err = strategyclientset.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(identitatemClientSet).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Process Strategy: ", func() {
	It("process a Strategy CR", func() {
		By("creating a Strategy CR", func() {
			strategy := identitatemv1alpha1.Strategy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mystrategy",
					Namespace: "default",
				},
				Spec: identitatemv1alpha1.StrategySpec{
					Type: identitatemv1alpha1.GrcStrategyType,
				},
			}
			_, err := identitatemClientSet.IdentityconfigV1alpha1().Strategies("default").Create(context.TODO(), &strategy, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		Eventually(func() error {
			r := StrategyReconciler{
				Client: k8sClient,
				Log:    logf.Log,
				Scheme: scheme.Scheme,
			}

			req := ctrl.Request{}
			req.Name = "mystrategy"
			req.Namespace = "default"
			_, err := r.Reconcile(context.TODO(), req)
			if err != nil {
				return err
			}
			authRealm, err := identitatemClientSet.IdentityconfigV1alpha1().Strategies("default").Get(context.TODO(), "mystrategy", metav1.GetOptions{})
			if err != nil {
				logf.Log.Info("Error while reading authrealm", "Error", err)
				return err
			}
			if len(authRealm.Spec.Type) == 0 {
				logf.Log.Info("StrategyType is still empty")
				return fmt.Errorf("Strategy %s/%s not processed", authRealm.Namespace, authRealm.Name)
			}
			return nil
		}, 30, 1).Should(BeNil())
	})
})
