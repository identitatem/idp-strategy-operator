// Copyright Red Hat

package api

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	identitatemv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Process Strategy: ", func() {
	var sampleStrategy identitatemv1alpha1.Strategy
	BeforeEach(func() {
		b, err := ioutil.ReadFile(filepath.Join("..", "config", "samples", "identitatem.io_v1alpha1_strategy.yaml"))
		Expect(err).ToNot(HaveOccurred())
		Expect(b).ShouldNot(BeNil())
		sampleStrategy = identitatemv1alpha1.Strategy{}
		err = yaml.Unmarshal(b, &sampleStrategy)
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		cr := sampleStrategy.DeepCopy()
		dynamicClient.Resource(identitatemv1alpha1.SchemeGroupVersion.WithResource("strategies")).
			Namespace(cr.Namespace).Delete(context.TODO(), cr.Name, metav1.DeleteOptions{})
	})
	It("create a Strategy CR", func() {
		cr := sampleStrategy.DeepCopy()
		createdCR, err := clientSet.IdentityconfigV1alpha1().Strategies(cr.Namespace).Create(context.TODO(), cr, metav1.CreateOptions{})
		Expect(err).To(BeNil())
		cu, err := dynamicClient.Resource(identitatemv1alpha1.SchemeGroupVersion.WithResource("strategies")).
			Namespace(cr.Namespace).
			Get(context.TODO(), cr.Name, metav1.GetOptions{})
		Expect(err).To(BeNil())
		c := &identitatemv1alpha1.Strategy{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(cu.UnstructuredContent(), c)
		Expect(err).To(BeNil())
		Expect(reflect.DeepEqual(createdCR.Spec, c.Spec)).To(BeTrue())
		Expect(reflect.DeepEqual(createdCR.ObjectMeta, c.ObjectMeta)).To(BeTrue())
	})
	It("read a Strategy CR", func() {
		cr := sampleStrategy.DeepCopy()
		content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
		cu := &unstructured.Unstructured{
			Object: content,
		}
		Expect(err).To(BeNil())
		_, err = dynamicClient.Resource(identitatemv1alpha1.SchemeGroupVersion.WithResource("strategies")).
			Namespace(cr.Namespace).Create(context.TODO(), cu, metav1.CreateOptions{})
		Expect(err).To(BeNil())
		c, err := clientSet.IdentityconfigV1alpha1().Strategies(cr.Namespace).Get(context.TODO(), sampleStrategy.Name, metav1.GetOptions{})
		Expect(err).Should(BeNil())
		Expect(reflect.DeepEqual(cr.Spec, c.Spec)).To(BeTrue())
		Expect(reflect.DeepEqual(cr.ObjectMeta.Name, c.ObjectMeta.Name)).To(BeTrue())
		Expect(reflect.DeepEqual(cr.ObjectMeta.Namespace, c.ObjectMeta.Namespace)).To(BeTrue())
	})
	It("delete a Strategy CR", func() {
		cr := sampleStrategy.DeepCopy()
		content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
		cu := &unstructured.Unstructured{
			Object: content,
		}
		Expect(err).To(BeNil())
		_, err = dynamicClient.Resource(identitatemv1alpha1.SchemeGroupVersion.WithResource("strategies")).
			Namespace(cr.Namespace).Create(context.TODO(), cu, metav1.CreateOptions{})
		Expect(err).To(BeNil())
		err = clientSet.IdentityconfigV1alpha1().Strategies(cr.Namespace).Delete(context.TODO(), sampleStrategy.Name, metav1.DeleteOptions{})
		Expect(err).Should(BeNil())
		err = dynamicClient.Resource(identitatemv1alpha1.SchemeGroupVersion.WithResource("strategies")).
			Namespace(cr.Namespace).Delete(context.TODO(), cr.Name, metav1.DeleteOptions{})
		Expect(err).ToNot(BeNil())
	})
})
