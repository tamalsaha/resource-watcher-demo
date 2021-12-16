/*
Copyright 2021.

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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/tamalsaha/resource-watcher-demo/vendor/k8s.io/klog/v2"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gomodules.xyz/sets"
	ksets "gomodules.xyz/sets/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2/klogr"
	"kmodules.xyz/authorizer/rbac"
	apiv1 "kmodules.xyz/client-go/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	gkSet    = ksets.NewGroupKind(
		schema.GroupKind{
			Group: "admissionregistration.k8s.io",
			Kind:  "ValidatingWebhookConfiguration",
		},
		schema.GroupKind{
			Group: "events.k8s.io",
			Kind:  "Event",
		},
		schema.GroupKind{
			Group: "storage.k8s.io",
			Kind:  "VolumeAttachment",
		},
		schema.GroupKind{
			Group: "admissionregistration.k8s.io",
			Kind:  "MutatingWebhookConfiguration",
		},
		schema.GroupKind{
			Group: "",
			Kind:  "PodTemplate",
		},
		schema.GroupKind{
			Group: "apps",
			Kind:  "ControllerRevision",
		},
		schema.GroupKind{
			Group: "apiextensions.k8s.io",
			Kind:  "CustomResourceDefinition",
		},
		schema.GroupKind{
			Group: "flowcontrol.apiserver.k8s.io",
			Kind:  "PriorityLevelConfiguration",
		},
		schema.GroupKind{
			Group: "",
			Kind:  "Event",
		})
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(klogr.New())

	cfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "783ac4f6.rswatcher.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	ctx := ctrl.SetupSignalHandler()

	//var ns v1.Namespace
	//err = mgr.GetClient().Get(context.TODO(), client.ObjectKey{Name: "default"}, &ns)
	//if err != nil {
	//	setupLog.Error(err, "unable to get namespace")
	//	// os.Exit(1)
	//}

	rbacAuthorizer := rbac.NewForManagerOrDie(ctx, mgr)
	fmt.Println(rbacAuthorizer)

	//if err := builder.ControllerManagedBy(mgr).For(&policyv1.PodDisruptionBudget{}).Complete(r); err != nil {
	//	panic(err)
	//}

	/*
		apiVersion: policy/v1
		kind: PodDisruptionBudget
		metadata:
		  name: zk-pdb
		spec:
		  minAvailable: 2
		  selector:
		    matchLabels:
		      app: zookeeper
	*/
	//mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
	//	// time.Sleep(1 * 30 * time.Second)
	//
	//	minA := intstr.FromInt(2)
	//	pdb1 := policyv1.PodDisruptionBudget{
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      "zk-pdb",
	//			Namespace: "default",
	//		},
	//		Spec: policyv1.PodDisruptionBudgetSpec{
	//			MinAvailable: &minA,
	//			Selector: &metav1.LabelSelector{
	//				MatchLabels: map[string]string{
	//					"app": "zookeeper",
	//				},
	//			},
	//		},
	//	}
	//	err = mgr.GetClient().Create(context.TODO(), &pdb1)
	//	if err != nil {
	//		setupLog.Error(err, "unable to create controller", "controller", "Release")
	//		os.Exit(1)
	//	}
	//
	//	var pdb2 policyv1beta1.PodDisruptionBudget
	//	err = mgr.GetClient().Get(context.TODO(), client.ObjectKey{Namespace: pdb1.Namespace, Name: pdb1.Name}, &pdb2)
	//	if err != nil {
	//		return err
	//	}
	//	fmt.Println(pdb2.Namespace + "/" + pdb2.Name)
	//	return nil
	//}))

	//if err = (&corecontrollers.ReleaseReconciler{
	//	Client: mgr.GetClient(),
	//	Scheme: mgr.GetScheme(),
	//}).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Release")
	//	os.Exit(1)
	//}
	////+kubebuilder:scaffold:builder

	resourceChannel := make(chan apiv1.ResourceID, 100)
	resourceTracker := map[schema.GroupVersionKind]apiv1.ResourceID{}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		_, h := setupGraphQL()
		http.Handle("/", h)
		log.Println("GraphQL running on port :8082")
		return http.ListenAndServe(":8082", nil)
	}))

	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		kc := kubernetes.NewForConfigOrDie(cfg)
		err := wait.PollImmediateUntil(60*time.Second, func() (done bool, err error) {
			rsLists, err := kc.Discovery().ServerPreferredResources()
			if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
				klog.ErrorS(err, "failed to list server preferred resources")
				return false, nil
			}
			for _, rsList := range rsLists {
				for _, rs := range rsList.APIResources {
					// skip sub resource
					if strings.ContainsRune(rs.Name, '/') {
						continue
					}

					// if resource can't be listed or read (get) skip it
					verbs := sets.NewString(rs.Verbs...)
					if !verbs.HasAll("list", "get", "watch") {
						continue
					}

					gvk := schema.FromAPIVersionAndKind(rsList.GroupVersion, rs.Kind)
					if gkSet.Has(gvk.GroupKind()) {
						continue
					}

					rid := apiv1.ResourceID{
						Group:   gvk.Group,
						Version: gvk.Version,
						Name:    rs.Name,
						Kind:    rs.Kind,
						Scope:   apiv1.ClusterScoped,
					}
					if rs.Namespaced {
						rid.Scope = apiv1.NamespaceScoped
					}
					if _, found := resourceTracker[gvk]; !found {
						resourceTracker[gvk] = rid
						resourceChannel <- rid
					}
				}
			}
			return false, nil
		}, ctx.Done())
		if err != nil {
			return err
		}

		close(resourceChannel)
		return nil
	}))

	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		for rid := range resourceChannel {
			if err := (&Reconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
				R:      rid,
			}).SetupWithManager(mgr); err != nil {
				return err
			}
			//var obj unstructured.Unstructured
			//obj.SetGroupVersionKind(rid.GroupVersionKind())
			//if err := builder.ControllerManagedBy(mgr).For(&obj).Complete(gr); err != nil {
			//	return err
			//}
		}
		return nil
	}))

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
