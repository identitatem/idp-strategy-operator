// Copyright Red Hat

// +build functional

package functional

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	identitatemclientset "github.com/identitatem/idp-client-api/api/client/clientset/versioned"
	identitatemv1alpha1 "github.com/identitatem/idp-client-api/api/identitatem/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)

}

var identitattemClientSet *identitatemclientset.Clientset
var cfg *rest.Config

var _ = Describe("Strategy", func() {
	BeforeEach(func() {
		logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter)))
		SetDefaultEventuallyTimeout(20 * time.Second)
		SetDefaultEventuallyPollingInterval(1 * time.Second)

		var err error
		kubeConfigFile := os.Getenv("KUBECONFIG")
		if len(kubeConfigFile) == 0 {
			home := homedir.HomeDir()
			kubeConfigFile = filepath.Join(home, ".kube", "config")
		}
		klog.Infof("KUBECONFIG=%s", kubeConfigFile)
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())
		identitattemClientSet, err = identitatemclientset.NewForConfig(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(identitattemClientSet).ToNot(BeNil())
	})

	AfterEach(func() {
	})

	It("process a Strategy", func() {
		name := "my-backplane-strategy"
		namespace := "default"

		By("Create a Backplane Strategy", func() {
			strategy := &identitatemv1alpha1.Strategy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: identitatemv1alpha1.StrategySpec{
					Type: identitatemv1alpha1.BackplaneStrategyType,
				},
			}
			_, err := identitattemClientSet.IdentityconfigV1alpha1().Strategies(namespace).Create(context.TODO(), strategy, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		Eventually(func() error {
			strategy, err := identitattemClientSet.IdentityconfigV1alpha1().Strategies(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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

		name = "my-grc-strategy"
		namespace = "default"
		By("Create a GRC Strategy", func() {

			strategy := identitatemv1alpha1.Strategy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: identitatemv1alpha1.StrategySpec{
					Type: identitatemv1alpha1.GrcStrategyType,
				},
			}
			_, err := identitattemClientSet.IdentityconfigV1alpha1().Strategies(namespace).Create(context.TODO(), &strategy, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		Eventually(func() error {
			strategy, err := identitattemClientSet.IdentityconfigV1alpha1().Strategies(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				logf.Log.Info("Error while reading strategy", "Error", err)
				return err
			}

			if len(strategy.Spec.Type) == 0 || strategy.Spec.Type != identitatemv1alpha1.GrcStrategyType {
				logf.Log.Info("Strategy Type is still empty")
				return fmt.Errorf("Strategy %s/%s not processed", strategy.Namespace, strategy.Name)
			}
			return nil
		}, 30, 1).Should(BeNil())

	})

})
