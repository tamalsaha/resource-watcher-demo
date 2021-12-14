package main

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2/klogr"
	clientutil "kmodules.xyz/client-go/client"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	_ = clientgoscheme.AddToScheme(scheme)

	ctrl.SetLogger(klogr.New())
	cfg := ctrl.GetConfigOrDie()

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return err
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		Opts: client.WarningHandlerOptions{
			SuppressWarnings:   false,
			AllowDuplicateLogs: false,
		},
	})
	if err != nil {
		return err
	}

	var nodes core.NodeList
	err = c.List(context.TODO(), &nodes)
	if err != nil {
		panic(err)
	}
	for _, n := range nodes.Items {
		fmt.Println(n.Name)
	}

	// Example:
	//var mysqls v1alpha2.MySQLList
	//err = c.List(context.TODO(), &mysqls)
	//if err != nil {
	//	panic(err)
	//}
	//for _, n := range mysqls.Items {
	//	fmt.Println(n.Name)
	//}

	/*
		apiVersion: v1
		kind: Pod
		metadata:
		  name: busybox
		  labels:
		    app: busybox
		spec:
		  containers:
		  - image: ubuntu:20.04
		    command:
		      - sleep
		      - "3600"
		    imagePullPolicy: IfNotPresent
		    name: busybox
		    resources:
		      limits:
		        cpu: 250m
		        memory: 256Mi
		  restartPolicy: Always
	*/
	finalObj, vt, err := clientutil.CreateOrPatch(c, &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "busybox",
		},
	}, func(obj client.Object, createOp bool) client.Object {
		p := obj.(*core.Pod)
		p.Labels = meta_util.OverwriteKeys(p.Labels, map[string]string{
			"app": "busybox",
		})
		if createOp {
			core_util.EnsureOwnerReference(&p.ObjectMeta, nil /*SET to actual OWNER*/)
		}

		// NEVER DO THIS
		//p.Spec = core.PodSpec{
		//
		//}

		p.Spec.Containers = core_util.UpsertContainer(p.Spec.Containers, core.Container{
			Name:    "busybox",
			Image:   "ubuntu:20.04",
			Command: []string{"sleep", "3600"},
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU:    resource.MustParse("250m"),
					core.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
			ImagePullPolicy: core.PullIfNotPresent,
		})
		p.Spec.RestartPolicy = core.RestartPolicyAlways

		return p
	})
	if err != nil {
		return err
	}
	fmt.Println("op = ", vt)
	fmt.Printf("obj = %+v", finalObj)

	return nil
}
