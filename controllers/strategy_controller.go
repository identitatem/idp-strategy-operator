// Copyright Red Hat

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	identitatemmgmtv1alpha1 "github.com/identitatem/idp-mgmt-operator/api/identitatem/v1alpha1"
	identitatemstrategyv1alpha1 "github.com/identitatem/idp-strategy-operator/api/identitatem/v1alpha1"

	//ocm "github.com/open-cluster-management-io/api/cluster/v1alpha1"
	clusterapiv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
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
	placementInfo := &identitatemmgmtv1alpha1.Placement{}
	//for _, or := range ownerRefs {
	for _, or := range instance.GetOwnerReferences() {
		r.Log.Info("Check OwnerRef ", or.Name)

		if or.Kind == authrealm.Kind {
			placementInfo = authrealm.Spec.Placement
			r.Log.Info("Found AuthRealm.  Placement name", placementInfo.Name)
			ownerRef = or
			break
		}
	}
	if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: ownerRef.Name, Namespace: req.Namespace}, authrealm); err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	authRealmPlacement := &clusterapiv1alpha1.Placement{}
	if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: placementInfo.Name, Namespace: req.Namespace}, authRealmPlacement); err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//Make sure Placement is created and correct
	//TODO!!! Right now we will have to manullay add a label to managed clusters in order for the placementDecision
	//        to return a result cloudservices=grc|backplane
	placementExists := true
	newPlacement := &clusterapiv1alpha1.Placement{}
	if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, newPlacement); err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		placementExists = false
		// Not Found! Create
		newPlacement = &clusterapiv1alpha1.Placement{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: req.Namespace,
				Name:      req.Name,
			},
			Spec: clusterapiv1alpha1.PlacementSpec{
				Predicates: []clusterapiv1alpha1.ClusterPredicate{
					{
						RequiredClusterSelector: clusterapiv1alpha1.ClusterSelector{
							LabelSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{
									"cloudservices": string(instance.Spec.Type),
								},
							},
						},
					},
				},
			},
		}
		// Append any additional predicates the AuthRealm already had on it's Placement
		newPlacement.Spec.Predicates = append(newPlacement.Spec.Predicates, authrealm.Spec.Placement.Spec.Predicates...)
		// Set owner reference for cleanup
		controllerutil.SetOwnerReference(instance, newPlacement, r.Scheme)
	}

	switch placementExists {
	case true:
		if err := r.Client.Update(context.TODO(), newPlacement); err != nil {
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}
	case false:
		if err := r.Client.Create(context.Background(), newPlacement); err != nil {
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}
	}

	// update the Placement ref
	instance.Spec.PlacementRef.Name = newPlacement.Name
	if err := r.Client.Update(context.Background(), instance); err != nil {
		// Error updating the object - requeue the request.
		return reconcile.Result{}, err
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
		Owns(&clusterapiv1alpha1.Placement{}).
		//Watches(&source.Kind{Type: &clusterapiv1alpha1.PlacementDecision{}, })  //TODO
		Complete(r)
}
