// Copyright Red Hat

package clusteroauth

import (
	"context"
	"encoding/json"
	"fmt"

	//"fmt"
	"reflect"

	//"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"github.com/identitatem/idp-client-api/api/client/clientset/versioned/scheme"
	identitatemv1alpha1 "github.com/identitatem/idp-client-api/api/identitatem/v1alpha1"

	// identitatemdexserverv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"
	identitatemdexv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"

	//ocm "github.com/open-cluster-management-io/api/cluster/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	idpconfig "github.com/identitatem/idp-client-api/config"
	openshiftconfigv1 "github.com/openshift/api/config/v1"

	//+kubebuilder:scaffold:imports

	clusteradmapply "open-cluster-management.io/clusteradm/pkg/helpers/apply"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// ClusterOAuthReconciler reconciles a Strategy object
type ClusterOAuthReconciler struct {
	client.Client
	KubeClient         kubernetes.Interface
	DynamicClient      dynamic.Interface
	APIExtensionClient apiextensionsclient.Interface
	Log                logr.Logger
	Scheme             *runtime.Scheme
}

var log = logf.Log.WithName("utils")

//+kubebuilder:rbac:groups=identityconfig.identitatem.io,resources={authrealms,strategies,clusteroauths},verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=identityconfig.identitatem.io,resources=strategies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=identityconfig.identitatem.io,resources=strategies/finalizers,verbs=update
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources={customresourcedefinitions},verbs=get;list;create;update;patch;delete

//+kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources={placements,placementdecisions},verbs=get;list;watch;create;update;patch;delete;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Strategy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ClusterOAuthReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("clusteroauth", req.NamespacedName)

	// your logic here
	// Fetch the ClusterOAuth instance
	instance := &identitatemv1alpha1.ClusterOAuth{}

	if err := r.Client.Get(
		context.TODO(),
		types.NamespacedName{Namespace: req.Namespace, Name: req.Name},
		instance,
	); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.Log.Info("Instance", "instance", instance)
	r.Log.Info("Running Reconcile for ClusterOAuth.", "Name: ", instance.GetName(), " Namespace:", instance.GetNamespace())

	//TODO   - I think this only applies to backplane so no need to check
	//         If grc also uses this, we need to have a way of knowing what strategy type created
	//         the CR so we can build ManifestWork properly for that strategy.  For example, only secrets in
	//         GRC manifestwork
	//switch instance.Spec.Type {
	//case identitatemv1alpha1.BackplaneStrategyType:

	// Create empty manifest work
	manifestWork := &manifestworkv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "idp-backplane",
			Namespace: instance.GetNamespace(),
		},
		Spec: manifestworkv1.ManifestWorkSpec{
			Workload: manifestworkv1.ManifestsTemplate{
				Manifests: []manifestworkv1.Manifest{},
			},
		},
	}

	// Get a list of all clusterOAuth
	clusterOAuths := &identitatemv1alpha1.ClusterOAuthList{}
	//	singleOAuth := &openshiftconfigv1.OAuth{}
	singleOAuth := &openshiftconfigv1.OAuth{
		TypeMeta: metav1.TypeMeta{
			APIVersion: openshiftconfigv1.SchemeGroupVersion.String(),
			Kind:       "OAuth",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      "idp-backplane-oauth",
			Namespace: instance.GetNamespace(),
		},

		Spec: openshiftconfigv1.OAuthSpec{},
	}

	if err := r.List(context.TODO(), clusterOAuths, &client.ListOptions{Namespace: instance.GetNamespace()}); err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	for _, clusterOAuth := range clusterOAuths.Items {
		//build OAuth and add to manifest work
		r.Log.Info("ClusterOAuth.", "Name: ", clusterOAuth.GetName(), " Namespace:", instance.GetNamespace(), "IdentityProviders:", len(instance.Spec.OAuth.Spec.IdentityProviders))

		for j, idp := range instance.Spec.OAuth.Spec.IdentityProviders {

			r.Log.Info("ClusterOAuth.", "IdentityProvider  ", j, " Name:", idp.Name)

			//build oauth by appending first clusterOAuth entry into single OAuth
			singleOAuth.Spec.IdentityProviders = append(singleOAuth.Spec.IdentityProviders, idp)

			//Look for secret for Identity Provider and if found, add to manifest work
			secret := &corev1.Secret{}

			if err := r.Client.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: idp.Name}, secret); err == nil {
				//add secret to manifest

				//TODO TEMP PATCH
				if len(secret.TypeMeta.Kind) == 0 {
					secret.TypeMeta.Kind = "Secret"

				}
				if len(secret.TypeMeta.APIVersion) == 0 {
					secret.TypeMeta.APIVersion = corev1.SchemeGroupVersion.String()

				}

				data, err := json.Marshal(secret)
				if err != nil {
					return reconcile.Result{}, err
				}

				manifest := manifestworkv1.Manifest{
					RawExtension: runtime.RawExtension{Raw: data},
				}

				//add manifest to manifest work
				manifestWork.Spec.Workload.Manifests = append(manifestWork.Spec.Workload.Manifests, manifest)

			}
		}
	}

	// create manifest for single OAuth
	data, err := json.Marshal(singleOAuth)
	if err != nil {
		return reconcile.Result{}, err
	}

	manifest := manifestworkv1.Manifest{
		RawExtension: runtime.RawExtension{Raw: data},
	}

	//add OAuth manifest to manifest work
	manifestWork.Spec.Workload.Manifests = append(manifestWork.Spec.Workload.Manifests, manifest)

	// create manifest work for managed cluster
	// (borrowed from https://github.com/open-cluster-management/endpoint-operator/blob/master/pkg/utils/utils.go)
	if err := CreateOrUpdateManifestWork(manifestWork, r.Client, manifestWork, r.Scheme); err != nil {
		r.Log.Error(err, "Failed to create manifest work for component")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//case identitatemv1alpha1..GRCStrategyType:
	//default:
	//	return reconcile.Result{}, fmt.Errorf("strategy type %s not supported", instance.Spec.Type)
	//
	//}
	return ctrl.Result{}, nil
}

// compareManifestWorks returns true if 2 manifestworks' specs are the same
func compareManifestWorks(mw1 *manifestworkv1.ManifestWork, mw2 *manifestworkv1.ManifestWork) bool {
	if mw1 == nil && mw2 == nil {
		return true
	}
	if (mw1 == nil && mw2 != nil) || (mw2 == nil && mw1 != nil) {
		return false
	}
	if len(mw1.Spec.Workload.Manifests) != len(mw2.Spec.Workload.Manifests) {
		return false
	}
	used := make(map[int]bool)
	for _, m1 := range mw1.Spec.Workload.Manifests {
		hasMatch := false
		for j, m2 := range mw2.Spec.Workload.Manifests {
			if used[j] {
				continue
			}
			if compareManifests(&m1.RawExtension, &m2.RawExtension) {
				hasMatch = true
				used[j] = true
				break
			}
		}
		if !hasMatch {
			return false
		}
	}
	return true
}

// convertRawExtensiontoUnstructured converts a rawExtension to a unstructured object
func convertRawExtensiontoUnstructured(r *runtime.RawExtension) (*unstructured.Unstructured, error) {
	if r == nil {
		return nil, fmt.Errorf("fail to convert rawExtension")
	}
	var obj runtime.Object
	var scope conversion.Scope
	err := runtime.Convert_runtime_RawExtension_To_runtime_Object(r, &obj, scope)
	if err != nil {
		log.Error(err, "failed to convert rawExtension to runtime.Object", "rawExtension", r)
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	innerObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		log.Error(err, "failed to convert runtime.Objectt to Unstructured", "runtime.Object", &obj)
		return nil, err
	}
	u := unstructured.Unstructured{Object: innerObj}
	return &u, nil
}

var rootAttributes = []string{
	"spec",
	"rules",
	"roleRef",
	"subjects",
	"secrets",
	"imagePullSecrets",
	"automountServiceAccountToken",
	"data",
}

// compareManifests compares if 2 manifests are the same, it only checks value we care
// (name/namespace/kind/group/spec/data)
func compareManifests(r1, r2 *runtime.RawExtension) bool {
	u1, err := convertRawExtensiontoUnstructured(r1)
	if err != nil {
		return false
	}
	u2, err := convertRawExtensiontoUnstructured(r2)
	if err != nil {
		return false
	}
	if u1 == nil || u2 == nil {
		return u2 == nil && u1 == nil
	}
	if u1.GetName() != u2.GetName() ||
		u1.GetNamespace() != u2.GetNamespace() ||
		u1.GetKind() != u2.GetKind() ||
		u1.GetAPIVersion() != u2.GetAPIVersion() {
		return false
	}
	hasDiff := false
	for _, r := range rootAttributes {
		if newValue, ok := u2.Object[r]; ok {
			if !reflect.DeepEqual(newValue, u1.Object[r]) {
				hasDiff = true
			}
		} else {
			if _, ok := u1.Object[r]; ok {
				hasDiff = true
			}
		}
	}
	return !hasDiff
}

// CreateOrUpdateManifestWork creates a new ManifestWork or update an existing ManifestWork
func CreateOrUpdateManifestWork(
	manifestwork *manifestworkv1.ManifestWork,
	client client.Client,
	owner metav1.Object,
	scheme *runtime.Scheme,
) error {
	var oldManifestwork manifestworkv1.ManifestWork

	err := client.Get(
		context.TODO(),
		types.NamespacedName{Name: manifestwork.Name, Namespace: manifestwork.Namespace},
		&oldManifestwork,
	)
	if err == nil {
		// Check if update is require
		if !compareManifestWorks(&oldManifestwork, manifestwork) {
			oldManifestwork.Spec.Workload.Manifests = manifestwork.Spec.Workload.Manifests
			if err := client.Update(context.TODO(), &oldManifestwork); err != nil {
				log.Error(err, "Fail to update manifestwork")
				return err
			}
		}
	} else {
		if errors.IsNotFound(err) {
			//if err := controllerutil.SetControllerReference(owner, manifestwork, scheme); err != nil {
			//	log.Error(err, "Unable to SetControllerReference")
			//	return err
			//}
			if err := client.Create(context.TODO(), manifestwork); err != nil {
				log.Error(err, "Fail to create manifestwork")
				return err
			}
			return nil
		}
		return err
	}

	return nil
}

// DeleteManifestWork deletes a manifestwork
// if removeFinalizers is set to true, will remove all finalizers to make sure it can be deleted
func DeleteManifestWork(name, namespace string, client client.Client, removeFinalizers bool) error {
	manifestWork := &manifestworkv1.ManifestWork{}
	var retErr error
	if err := client.Get(
		context.TODO(),
		types.NamespacedName{Name: name, Namespace: namespace},
		manifestWork,
	); err != nil {
		return err
	}

	if removeFinalizers && len(manifestWork.GetFinalizers()) > 0 {
		manifestWork.SetFinalizers([]string{})
		if err := client.Update(context.TODO(), manifestWork); err != nil {
			log.Error(err, fmt.Sprintf("Failed to remove finalizers of Manifestwork %s in %s namespace", name, namespace))
			retErr = err
		}
	}

	if manifestWork.DeletionTimestamp == nil {
		err := client.Delete(context.TODO(), manifestWork)
		if err != nil {
			return err
		}
	}

	return retErr
}

func GetManifestWork(name, namespace string, client client.Client) (*manifestworkv1.ManifestWork, error) {
	manifestWork := &manifestworkv1.ManifestWork{}

	if err := client.Get(
		context.TODO(),
		types.NamespacedName{Name: name, Namespace: namespace},
		manifestWork,
	); err != nil {
		return nil, err
	}

	return manifestWork, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterOAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//Install CRD
	applierBuilder := &clusteradmapply.ApplierBuilder{}
	applier := applierBuilder.WithClient(r.KubeClient, r.APIExtensionClient, r.DynamicClient).Build()

	readerIDPMgmtOperator := idpconfig.GetScenarioResourcesReader()

	file := "crd/bases/identityconfig.identitatem.io_clusteroauths.yaml"
	if _, err := applier.ApplyDirectly(readerIDPMgmtOperator, nil, false, "", file); err != nil {
		return err
	}

	if err := corev1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := identitatemv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := clusterv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := clusterv1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := manifestworkv1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := identitatemdexv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	if err := openshiftconfigv1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&identitatemv1alpha1.ClusterOAuth{}).
		Complete(r)
}
