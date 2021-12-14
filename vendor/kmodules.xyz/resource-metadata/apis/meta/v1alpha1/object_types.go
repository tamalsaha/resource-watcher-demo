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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func ParseObjectID(s string) (*ObjectID, error) {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '='
	})

	var id ObjectID
	for i := 0; i < len(parts); i += 2 {
		switch parts[i] {
		case "G":
			id.Group = parts[i+1]
		case "K":
			id.Kind = parts[i+1]
		case "NS":
			id.Namespace = parts[i+1]
		case "N":
			id.Name = parts[i+1]
		default:
			return nil, fmt.Errorf("unknown key %s", parts[i])
		}
	}
	return &id, nil
}

func (oid ObjectID) GroupKind() schema.GroupKind {
	return schema.GroupKind{Group: oid.Group, Kind: oid.Kind}
}

func (oid ObjectID) ObjectKey() client.ObjectKey {
	return client.ObjectKey{Namespace: oid.Namespace, Name: oid.Name}
}
