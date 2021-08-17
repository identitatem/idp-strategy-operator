// Copyright Red Hat

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	identitatemmgmtv1alpha1 "github.com/identitatem/idp-mgmt-operator/api/identitatem/v1alpha1"
	identitatemstrategyv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"
	//ocm "github.com/open-cluster-management-io/api/cluster/v1alpha1"
)

// StrategyReconciler reconciles a Strategy object
type StrategyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
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

	// Check for placement and build if it does not exist
	///TODO Why can't I check for nil?
	if instance.Spec.PlacementRef.Size() == 0 {
		authrealm := &identitatemmgmtv1alpha1.AuthRealm{}
		//placementInfo := &identitatemmgmtv1alpha1.Placement{}
		//var predicates []clusterapiv1alpha1.ClusterPredicate
		//matchexpressions := &clusterapiv1alpha1.ClusterClaimSelector.MatchExpressions[]
		//labelselector := &metav1.LabelSelector.

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

		//	newPlacement := &clusterapiv1alpha1.Placement{
		//			ObjectMeta: metav1.ObjectMeta{
		//				Namespace: req.Namespace,
		//				Name:      req.Name,
		//			},
		//			Spec: clusterapiv1alpha1.PlacementSpec{
		//				Predicates: predicates,
		//			},
		//		}
		//placement, err = clusterClient.ClusterV1alpha1().Placements(namespace).Create(context.Background(), newPlacement, metav1.CreateOptions{})

		//clusterapiv1alpha1.PlacementDecision

		// get placement info from ownerRef AuthRealm
		ownerRefs := instance.GetOwnerReferences()
		//ownerRefs := instance.ObjectMeta.OwnerReferences
		//ownerRef := &metav1.OwnerReference{}

		for _, ownerRef := range ownerRefs {
			if ownerRef.Kind == authrealm.Kind {
				placementInfo := authrealm.Spec.Placement
				r.Log.Info("Placement name", placementInfo.Name)
			}
		}
	}
	// Check StrategyType
	//Backplane/Multi cluster engine for Kubernetes
	if instance.Spec.Type == identitatemstrategyv1alpha1.BackplaneStrategyType {
		r.Log.Info("Instance", "Type", instance.Spec.Type)

		// create placement with backplane filters/predicates
		//   add predicate for policy controller addon NOT PRESENT
		// condition = PlacementCreated
		// need to wait for PlacementDecision to trigger next step

		// GRC available - Advanced Cluster Management
	} else if instance.Spec.Type == identitatemstrategyv1alpha1.GrcStrategyType {

		r.Log.Info("Instance", "Type", instance.Spec.Type)

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

	} else {
		r.Log.Info("Instance", "StrategyType", instance.Spec.Type)
		return reconcile.Result{}, nil
	}

	if err := r.Client.Update(context.TODO(), instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StrategyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&identitatemstrategyv1alpha1.Strategy{}).
		Complete(r)
}
