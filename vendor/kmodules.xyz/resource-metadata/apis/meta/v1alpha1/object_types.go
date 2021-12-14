/*
Copyright AppsCode Inc. and Contributors

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

package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectLocator struct {
	Src  ObjectRef `json:"src"`
	Path []string  `json:"path"` // sequence of DirectedEdge names
}

type NamedEdge struct {
	Name       string                 `json:"name"`
	Src        metav1.TypeMeta        `json:"src"`
	Dst        metav1.TypeMeta        `json:"dst"`
	Connection ResourceConnectionSpec `json:"connection"`
}

type ObjectRef struct {
	Target    metav1.TypeMeta       `json:"target"`
	Selector  *metav1.LabelSelector `json:"selector,omitempty"`
	Name      string                `json:"name,omitempty"`
	Namespace string                `json:"namespace,omitempty"`
}

type ObjectID struct {
	Group     string `json:"group,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func (oid ObjectID) String() string {
	return fmt.Sprintf("G=%s,K=%s,NS=%s,N=%s", oid.Group, oid.Kind, oid.Namespace, oid.Name)
}

func NewObjectID(obj client.Object) ObjectID {
	gvk := obj.GetObjectKind().GroupVersionKind()
	return ObjectID{
		Group:     gvk.Group,
		Kind:      gvk.Kind,
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}
