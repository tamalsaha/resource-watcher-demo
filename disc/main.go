package main

import (
	"encoding/json"
	"fmt"
	"gomodules.xyz/sets"
	ksets "gomodules.xyz/sets/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	apiv1 "kmodules.xyz/client-go/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

func main() {
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
	rsLists, err := kc.Discovery().ServerPreferredResources()
	if err != nil {
		panic(err)
	}

	//result, err := restmapper.GetAPIGroupResources(kc.Discovery())
	//if err != nil {
	//	panic(err)
	//}
	//data, err := json.MarshalIndent(result, "", "  ")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(string(data))

	m2 := map[schema.GroupVersionKind]metav1.APIResource{}
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

			rs.Group = gvk.Group
			rs.Version = gvk.Version
			m2[gvk] = rs
		}
	}

	result := make([]apiv1.ResourceID, 0, len(m2))
	for _, rs := range m2 {
		rid := apiv1.ResourceID{
			Group:   rs.Group,
			Version: rs.Version,
			Name:    rs.Name,
			Kind:    rs.Kind,
			Scope:   apiv1.ClusterScoped,
		}
		if rs.Namespaced {
			rid.Scope = apiv1.NamespaceScoped
		}
		result = append(result, rid)
	}

	data2, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data2))
}
