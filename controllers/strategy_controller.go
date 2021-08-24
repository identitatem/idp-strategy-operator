// Copyright Red Hat

package controllers

import (
	"context"
	"fmt"
	"time"

	ocinfrav1 "github.com/openshift/api/config/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	identitatemmgmtv1alpha1 "github.com/identitatem/idp-mgmt-operator/api/identitatem/v1alpha1"
	"github.com/identitatem/idp-strategy-operator/api/client/clientset/versioned/scheme"
	identitatemstrategyv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"

	// identitatemdexserverv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"
	identitatemdexv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"

	//ocm "github.com/open-cluster-management-io/api/cluster/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
	workv1 "open-cluster-management.io/api/work/v1"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	idpstrategyoperatorconfig "github.com/identitatem/idp-strategy-operator/config"

	//+kubebuilder:scaffold:imports

	clusteradmapply "open-cluster-management.io/clusteradm/pkg/helpers/apply"
)

// StrategyReconciler reconciles a Strategy object
type StrategyReconciler struct {
	client.Client
	KubeClient         kubernetes.Interface
	DynamicClient      dynamic.Interface
	APIExtensionClient apiextensionsclient.Interface
	Log                logr.Logger
	Scheme             *runtime.Scheme
}

//+kubebuilder:rbac:groups=identityconfig.identitatem.io,resources=strategies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=identityconfig.identitatem.io,resources=strategies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=identityconfig.identitatem.io,resources=strategies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Strategy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *StrategyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here
	// Fetch the ManagedCluster instance
	instance := &identitatemstrategyv1alpha1.Strategy{}

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

	r.Log.Info("Running Reconcile for Strategy.", "Name: ", instance.Name, " Namespace:", instance.Namespace)

	//apiVersion: cluster.open-cluster-management.io/v1alpha1
	//kind: Placement
	//metadata:
	//  name: placement-policy-cert-ocp4
	//spec:
	//  predicates:
	//  - requiredClusterSelector:
	//      labelSelector:
	//        matchExpressions:
	//          - {key: vendor, operator: In, values: ["OpenShift"]}

	//					// add a predicates
	//					placement, err := clusterClient.ClusterV1alpha1().Placements(namespace).Get(context.Background(), placementName, metav1.GetOptions{})
	//					placement.Spec.Predicates = []clusterapiv1alpha1.ClusterPredicate{
	//						{
	//							RequiredClusterSelector: clusterapiv1alpha1.ClusterSelector{
	//								LabelSelector: metav1.LabelSelector{
	//									MatchLabels: map[string]string{
	//										"cloudservices": instance.Spec.type,    //use for now until we get new function
	//									},
	//								},
	//							},
	//						},
	//					}
	//					placement, err = clusterClient.ClusterV1alpha1().Placements(namespace).Update(context.Background(), placement, metav1.UpdateOptions{})

	// Get the AuthRealm Placement bits we need to help create a new Placement
	authrealm := &identitatemmgmtv1alpha1.AuthRealm{}

	r.Log.Info("Looking for AuthRealm in ownerRefs")
	// get placement info from AuthRealm ownerRef
	var ownerRef metav1.OwnerReference
	//DV not needed
	// placementInfo := &identitatemmgmtv1alpha1.Placement{}

	//for _, or := range ownerRefs {
	for _, or := range instance.GetOwnerReferences() {
		//DV add a parameter as it should be key/value pair
		r.Log.Info("Check OwnerRef ", "name", or.Name)

		//TODO find a better way
		if or.Kind == "AuthRealm" {
			// placementInfo = authrealm.Spec.Placement
			// r.Log.Info("Found AuthRealm.  Placement name", placementInfo.Name)
			ownerRef = or
			break
		}
	}
	if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: ownerRef.Name, Namespace: req.Namespace}, authrealm); err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//DV Not needed
	// authRealmPlacement := &clusterapiv1alpha1.Placement{}
	// if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: placementInfo.Name, Namespace: req.Namespace}, authRealmPlacement); err != nil {
	// 	// Error reading the object - requeue the request.
	// 	return reconcile.Result{}, err
	// }

	//Make sure Placement is created and correct
	//TODO!!! Right now we will have to manullay add a label to managed clusters in order for the placementDecision
	//        to return a result cloudservices=grc|backplane
	//Check if there is a predicate to add, if not nothing to do

	placement := &clusterv1alpha1.Placement{}
	if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: authrealm.Spec.PlacementRef.Name, Namespace: req.Namespace}, placement); err != nil {
		return reconcile.Result{}, err
	}

	//Get placementStrategy
	placementStrategy, placementStrategyExists, err := r.getStrategyPlacement(instance, authrealm, placement)

	//Enrich placementStrategy
	switch instance.Spec.Type {
	case identitatemstrategyv1alpha1.BackplaneStrategyType:
		if err := r.backplanePlacementStrategy(instance, authrealm, placement, placementStrategy); err != nil {
			return reconcile.Result{}, err
		}
	case identitatemstrategyv1alpha1.GrcStrategyType:
		if err := r.grcPlacementStrategy(instance, authrealm, placement, placementStrategy); err != nil {
			return reconcile.Result{}, err
		}
	default:
		return reconcile.Result{}, fmt.Errorf("strategy type %s not supported", instance.Spec.Type)
	}

	//Create or update placementStrategy
	switch placementStrategyExists {
	case true:
		if err := r.Client.Update(context.TODO(), placementStrategy); err != nil {
			return reconcile.Result{}, err
		}
	case false:
		if err := r.Client.Create(context.Background(), placementStrategy); err != nil {
			return reconcile.Result{}, err
		}
	}

	// update the Placement ref
	instance.Spec.PlacementRef.Name = placement.Name
	if err := r.Client.Update(context.Background(), instance); err != nil {
		// Error updating the object - requeue the request.
		return reconcile.Result{}, err
	}

	//DV Check if PlacementDecision Available
	platecementDecision := &clusterv1alpha1.PlacementDecision{}
	err = r.Client.Get(context.TODO(), client.ObjectKey{Name: getPlacementStrategyName(instance, authrealm), Namespace: req.Namespace}, platecementDecision)
	if err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		//Wait 10 sec
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 10}, nil
	}

	//DV Use switch as it is nicer than if then else
	switch instance.Spec.Type {
	case identitatemstrategyv1alpha1.BackplaneStrategyType:
		if err := r.backplaneStrategy(instance, authrealm, placement, platecementDecision); err != nil {
			return reconcile.Result{}, err
		}
	case identitatemstrategyv1alpha1.GrcStrategyType:
		if err := r.grcStrategy(instance, placement, platecementDecision); err != nil {
			return reconcile.Result{}, err
		}
	default:
		return reconcile.Result{}, fmt.Errorf("strategy type %s not supported", instance.Spec.Type)
	}

	// Check StrategyType
	//Backplane/Multi cluster engine for Kubernetes
	// if instance.Spec.Type == identitatemstrategyv1alpha1.BackplaneStrategyType {
	// 	r.Log.Info("Instance", "Type", instance.Spec.Type)

	// create placement with backplane filters/predicates
	//   add predicate for policy controller addon NOT PRESENT
	// condition = PlacementCreated
	// need to wait for PlacementDecision to trigger next step

	// GRC available - Advanced Cluster Management
	// } else if instance.Spec.Type == identitatemstrategyv1alpha1.GrcStrategyType {

	// 	r.Log.Info("Instance", "Type", instance.Spec.Type)
	//		if instance.Status.Conditions == nil {
	//			var newConditions []metav1.Condition
	//			now := metav1.Now()
	//			newCondition := metav1.Condition{
	//				Type:               "TestStrategyType",
	//				Status:             metav1.ConditionUnknown,
	//				Reason:             "TestStrategyReason",
	//				Message:            "Condition Initialized",
	//				LastTransitionTime: now,
	//			}
	//			//newConditions.add(newCondition)
	//			newConditions = append(newConditions, newCondition)
	//			//newConditions[0] = newCondition
	//
	//			r.Log.Info("UpdateCondition")
	//
	//			instance.Status.Conditions = newConditions
	//			if err := r.Client.Update(context.TODO(), instance); err != nil {
	//				return ctrl.Result{}, err
	//			}
	//		}

	// GRC available - Advanced Cluster Management

	// get placement info from ownerRef AuthRealm
	// create placement with backplane filters/predicates
	//   add predicate for policy controller addon == true
	// condition = PlacementCreated
	// need to wait for PlacementDecision to trigger next step

	// } else {
	// 	r.Log.Info("Instance", "StrategyType", instance.Spec.Type)
	// 	return reconcile.Result{}, nil
	// }

	if err := r.Client.Update(context.TODO(), instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *StrategyReconciler) getStrategyPlacement(strategy *identitatemstrategyv1alpha1.Strategy,
	authrealm *identitatemmgmtv1alpha1.AuthRealm,
	placement *clusterv1alpha1.Placement) (*clusterv1alpha1.Placement, bool, error) {
	placementStrategy := &clusterv1alpha1.Placement{}
	placementStrategyExists := true
	placementStrategyName := getPlacementStrategyName(strategy, authrealm)
	if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: placementStrategyName, Namespace: strategy.Namespace}, placementStrategy); err != nil {
		if !errors.IsNotFound(err) {
			return nil, false, err
		}
		placementStrategyExists = false
		// Not Found! Create
		placementStrategy = &clusterv1alpha1.Placement{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: strategy.Namespace,
				//DV The name is given by the authrealm as the user will define the binding with the clusterset
				//Name:      req.Name,
				Name: placementStrategyName,
			},
			//DV move below
			Spec: placement.Spec,
		}
		// Set owner reference for cleanup
		controllerutil.SetOwnerReference(strategy, placementStrategy, r.Scheme)
	}
	return placementStrategy, placementStrategyExists, nil
}

func getPlacementStrategyName(strategy *identitatemstrategyv1alpha1.Strategy,
	authrealm *identitatemmgmtv1alpha1.AuthRealm) string {
	return fmt.Sprintf("%s-%s", authrealm.Spec.PlacementRef.Name, strategy.Spec.Type)
}

// SetupWithManager sets up the controller with the Manager.
func (r *StrategyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//Install CRD
	applierBuilder := &clusteradmapply.ApplierBuilder{}
	applier := applierBuilder.WithClient(r.KubeClient, r.APIExtensionClient, r.DynamicClient).Build()

	readerIDPMgmtOperator := idpstrategyoperatorconfig.GetScenarioResourcesReader()

	file := "crd/bases/identityconfig.identitatem.io_strategies.yaml"
	if _, err := applier.ApplyDirectly(readerIDPMgmtOperator, nil, false, "", file); err != nil {
		return err
	}

	if err := clusterv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := clusterv1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := workv1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := identitatemdexv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	if err := ocinfrav1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&identitatemstrategyv1alpha1.Strategy{}).
		Owns(&clusterv1alpha1.Placement{}).
		Watches(&source.Kind{Type: &clusterv1alpha1.PlacementDecision{}},
			handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
				//Search the placement corresponding to the placementDecision
				placement := &clusterv1alpha1.Placement{}
				err := mgr.GetClient().Get(context.TODO(),
					client.ObjectKey{
						Name:      o.GetName(),
						Namespace: o.GetNamespace(),
					}, placement)
				if err != nil {
					r.Log.Error(err, "Error while getting placement")
					return []ctrl.Request{}
				}
				//Search the strategies for that placement
				strategies := &identitatemstrategyv1alpha1.StrategyList{}
				err = r.Client.List(context.TODO(), strategies, client.MatchingFields{
					"spec.placementRef.name": placement.Name,
				}, &client.ListOptions{Namespace: o.GetNamespace()})
				if err != nil {
					r.Log.Error(err, "Error while getting the list of strategies")
					return []ctrl.Request{}
				}
				requests := make([]reconcile.Request, 0)
				for _, strategy := range strategies.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      strategy.Name,
							Namespace: strategy.Namespace,
						},
					})
				}
				return requests
			})).
		Complete(r)
}
