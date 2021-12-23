package main

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	apiv1 "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	"kmodules.xyz/resource-metadata/hub"
	ksets "kmodules.xyz/sets"
)

var reg = hub.NewRegistryOfKnownResources()

var objGraph = &ObjectGraph{
	m:     sync.RWMutex{},
	edges: map[apiv1.OID]map[v1alpha1.EdgeLabel]ksets.OID{},
	ids:   map[apiv1.OID]map[v1alpha1.EdgeLabel]ksets.OID{},
}

var Schema = getGraphQLSchema()

var resourceChannel = make(chan apiv1.ResourceID, 100)
var resourceTracker = map[schema.GroupVersionKind]apiv1.ResourceID{}

var gkSet = ksets.NewGroupKind(
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
