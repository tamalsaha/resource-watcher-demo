package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/graphql-go/handler"
	"github.com/tamalsaha/resource-watcher-demo/graph"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2/klogr"
	"kmodules.xyz/authorizer/rbac"
	apiv1 "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
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
	cfg.QPS = 100
	cfg.Burst = 100
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "783ac4f6.rswatcher.dev",
		ClientDisableCacheFor: []client.Object{
			&core.Pod{},
		},
		NewClient: func(cache cache.Cache, config *restclient.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
			c, err := client.New(config, options)
			if err != nil {
				return nil, err
			}

			return client.NewDelegatingClient(client.NewDelegatingClientInput{
				CacheReader:       cache,
				Client:            c,
				UncachedObjects:   uncachedObjects,
				CacheUnstructured: true,
			})
		},
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

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		h := handler.New(&handler.Config{
			Schema:     &graph.Schema,
			Pretty:     true,
			GraphiQL:   false,
			Playground: true,
		})

		http.Handle("/", h)
		http.Handle("/graph", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			objid, _ := apiv1.ParseObjectID("G=apps,K=Deployment,NS=kube-system,N=coredns")
			resp, err := graph.ResourceGraph(mgr.GetRESTMapper(), *objid)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintf(w, "failed to execute graphql operation, errors: %v", err)
				return
			}

			rJSON, _ := json.MarshalIndent(resp, "", "  ")
			w.Write(rJSON)
			return
		}))
		http.Handle("/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Query
			query := `query Find($src: String!, $targetGroup: String!, $targetKind: String!) {
  find(oid: $src) {
    refs: exposed_by(group: $targetGroup, kind: $targetKind) {
      namespace
      name
    }
  }
}`
			vars := map[string]interface{}{
				// v1alpha1.GraphQueryVarSource:      "G=apps,K=Deployment,NS=kube-system,N=calico-kube-controllers",
				// v1alpha1.GraphQueryVarSource:      "G=,K=Pod,NS=kube-system,N=coredns-64897985d-kcr42",
				v1alpha1.GraphQueryVarSource:      "G=,K=Deployment,NS=kube-system,N=coredns",
				v1alpha1.GraphQueryVarTargetGroup: "",
				v1alpha1.GraphQueryVarTargetKind:  "Service",
				//v1alpha1.GraphQueryVarTargetGroup: "apps",
				//v1alpha1.GraphQueryVarTargetKind:  "ReplicaSet",
			}
			objs, err := graph.ExecGraphQLQuery(mgr.GetClient(), query, vars)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintf(w, "failed to execute graphql operation, errors: %v", err)
				return
			}

			oids := make([]apiv1.OID, 0, len(objs))
			for _, obj := range objs {
				oids = append(oids, apiv1.NewObjectID(&obj).OID())
			}
			rJSON, _ := json.MarshalIndent(oids, "", "  ")
			w.Write(rJSON)
			return
		}))
		log.Println("GraphQL running on port :8082")
		return http.ListenAndServe(":8082", nil)
	}))

	if err := mgr.Add(manager.RunnableFunc(graph.PollNewResourceTypes(cfg))); err != nil {
		setupLog.Error(err, "unable to set up resource poller")
		os.Exit(1)
	}

	if err := mgr.Add(manager.RunnableFunc(graph.SetupGraphReconciler(mgr))); err != nil {
		setupLog.Error(err, "unable to set up resource reconciler configurator")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
