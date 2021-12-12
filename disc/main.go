package main

import (
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

func main() {
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
			rs.Group = gvk.Group
			rs.Version = gvk.Version
			m2[gvk] = rs
		}
	}

	result := make([]metav1.APIResource, 0, len(m2))
	for _, rs := range m2 {
		result = append(result, rs)
	}

	data2, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data2))
}
