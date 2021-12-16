package main

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
	"kmodules.xyz/client-go/discovery"
	"kmodules.xyz/resource-metadata/hub"
	"kmodules.xyz/resource-metadata/pkg/graph"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiv1 "kmodules.xyz/client-go/api/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var reg = hub.NewRegistryOfKnownResources()

var objGraph = &ObjectGraph{
	m:     sync.RWMutex{},
	edges: map[string]sets.String{},
	ids:   map[string]sets.String{},
}

// Reconciler reconciles a Release object
type Reconciler struct {
	client.Client
	R      apiv1.ResourceID
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Release object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx).WithValues("name", req.NamespacedName.Name)
	gvk := r.R.GroupVersionKind()

	var obj unstructured.Unstructured
	obj.SetGroupVersionKind(gvk)
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		log.Error(err, "unable to fetch CronJob")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if rd, err := reg.LoadByGVK(gvk); err == nil {
		finder := graph.ObjectFinder{
			Client: r.Client,
			Mapper: discovery.NewResourceMapper(r.RESTMapper()),
		}
		if result, err := finder.ListConnectedObjectIDs(&obj, rd.Spec.Connections); err != nil {
			log.Error(err, "unable to fetch CronJob")
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, client.IgnoreNotFound(err)
		} else {
			objGraph.Update(apiv1.NewObjectID(&obj).Key(), result)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	var obj unstructured.Unstructured
	obj.SetGroupVersionKind(r.R.GroupVersionKind())
	return ctrl.NewControllerManagedBy(mgr).
		For(&obj).
		Complete(r)
}
