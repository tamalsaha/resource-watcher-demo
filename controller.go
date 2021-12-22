/*
Copyright AppsCode Inc. and Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package graph

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	apiv1 "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/resource-metadata/pkg/graph"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler reconciles a Release object
type Reconciler struct {
	client.Client
	R      apiv1.ResourceID
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx).WithValues("name", req.NamespacedName.Name)
	gvk := r.R.GroupVersionKind()

	var obj unstructured.Unstructured
	obj.SetGroupVersionKind(gvk)
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		log.Error(err, "unable to fetch", "group", r.R.Group, "kind", r.R.Kind)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if rd, err := reg.LoadByGVK(gvk); err == nil {
		finder := graph.ObjectFinder{
			Client: r.Client,
		}
		if result, err := finder.ListConnectedObjectIDs(&obj, rd.Spec.Connections); err != nil {
			log.Error(err, "unable to list connections", "group", r.R.Group, "kind", r.R.Kind)
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, client.IgnoreNotFound(err)
		} else {
			objGraph.Update(apiv1.NewObjectID(&obj).OID(), result)
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
