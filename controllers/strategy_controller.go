// Copyright Contributors to the Open Cluster Management project

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
	identitatemv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"
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
	instance := &identitatemv1alpha1.Strategy{}

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

	// Check StrategyType
	if instance.Spec.StrategyType == identitatemv1alpha1.BackplaneStrategyType {
		r.Log.Info("Instance", "StrategyType", instance.Spec.StrategyType)

		// create placement with backplane filters/predicates
		// condition = PlacementCreated

	} else if instance.Spec.StrategyType == identitatemv1alpha1.GrcStrategyType {

		r.Log.Info("Instance", "StrategyType", instance.Spec.StrategyType)

		// create placement with grc filters/predicaes
		// condition = PlacementCreated

	} else {
		r.Log.Info("Instance", "StrategyType", instance.Spec.StrategyType)
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
		For(&identitatemv1alpha1.Strategy{}).
		Complete(r)
}
