package main

import (
	"fmt"
	"strings"
	"time"

	"gomodules.xyz/sets"
	ksets "gomodules.xyz/sets/kubernetes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	apiv1 "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/resource-metadata/pkg/graph"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	g, err := graph.LoadGraphOfKnownResources()
	if err != nil {
		panic(err)
	}
	fmt.Println(g)

	gkSet := ksets.NewGroupKind(
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

	cfg := ctrl.GetConfigOrDie()
	kc := kubernetes.NewForConfigOrDie(cfg)
	//rsLists, err := kc.Discovery().ServerPreferredResources()
	//if err != nil {
	//	panic(err)
	//}

	//result, err := restmapper.GetAPIGroupResources(kc.Discovery())
	//if err != nil {
	//	panic(err)
	//}
	//data, err := json.MarshalIndent(result, "", "  ")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(string(data))

	resourceChannel := make(chan apiv1.ResourceID, 100)
	resourceTracker := map[schema.GroupVersionKind]apiv1.ResourceID{}
	//for _, rsList := range rsLists {
	//	for _, rs := range rsList.APIResources {
	//		// skip sub resource
	//		if strings.ContainsRune(rs.Name, '/') {
	//			continue
	//		}
	//
	//		// if resource can't be listed or read (get) skip it
	//		verbs := sets.NewString(rs.Verbs...)
	//		if !verbs.HasAll("list", "get", "watch") {
	//			continue
	//		}
	//
	//		gvk := schema.FromAPIVersionAndKind(rsList.GroupVersion, rs.Kind)
	//		if gkSet.Has(gvk.GroupKind()) {
	//			continue
	//		}
	//
	//		rid := apiv1.ResourceID{
	//			Group:   gvk.Group,
	//			Version: gvk.Version,
	//			Name:    rs.Name,
	//			Kind:    rs.Kind,
	//			Scope:   apiv1.ClusterScoped,
	//		}
	//		if rs.Namespaced {
	//			rid.Scope = apiv1.NamespaceScoped
	//		}
	//		if _, found := resourceTracker[gvk]; !found {
	//			resourceTracker[gvk] = rid
	//			resourceChannel <- rid
	//		}
	//	}
	//}
	//
	//result := make([]apiv1.ResourceID, 0, len(resourceTracker))
	//for _, rs := range resourceTracker {
	//	result = append(result, rs)
	//}
	//
	//data2, err := json.MarshalIndent(result, "", "  ")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(string(data2))

	go func() {
		for rid := range resourceChannel {
			fmt.Println(rid.GroupVersionKind())
		}
	}()

	err = wait.PollImmediateUntil(60*time.Second, func() (done bool, err error) {
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
	}, nil)
	if err != nil {
		panic(err)
	}

	select {}
}
