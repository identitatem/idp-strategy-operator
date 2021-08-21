module github.com/identitatem/idp-strategy-operator

go 1.16

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.4.0
	github.com/identitatem/dex-operator v0.0.4-0.20210818160217-d91d9387dda1
	github.com/identitatem/idp-mgmt-operator v0.0.0-20210816175333-8714ccda628f
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/openshift/api v0.0.0-20210521075222-e273a339932a
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.22.0
	k8s.io/klog/v2 v2.10.0
	open-cluster-management.io/api v0.0.0-20210804091127-340467ff6239
	open-cluster-management.io/clusteradm v0.1.0-alpha.5
	sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/controller-tools v0.5.0
)

replace (
	github.com/identitatem/idp-mgmt-operator => /Users/dvernier/go/src/github.com/identitatem/idp-mgmt-operator
	k8s.io/api => k8s.io/api v0.22.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.0
	k8s.io/client-go => k8s.io/client-go v0.22.0
	k8s.io/code-generator => k8s.io/code-generator v0.22.0
)
