package main

import (
	"fmt"
	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	"kmodules.xyz/resource-metadata/hub"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "kmodules.xyz/client-go/api/v1"
)

// https://github.com/graphql-go/graphql/blob/master/examples/star-wars/main.go
func setupGraphQL() (*graphql.Schema, http.Handler) {
	var (
		oidType *graphql.Object
	)

	oidType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "ObjectID",
		Description: "Uniquely identifies a Kubernetes object",
		Fields: graphql.Fields{
			"group": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The group of the Object",
			},
			"kind": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The kind of the Object",
			},
			"namespace": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The namespace of the Object",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The name of the Object.",
			},
		},
	})
	for _, label := range hub.ListEdgeLabels() {
		func(edgeLabel v1alpha1.EdgeLabel) {
			oidType.AddFieldConfig(string(edgeLabel), &graphql.Field{
				Type:        graphql.NewList(oidType),
				Description: fmt.Sprintf("%s from this object", edgeLabel),
				Args: graphql.FieldConfigArgument{
					"group": &graphql.ArgumentConfig{
						Description: "group of the linked objects",
						Type:        graphql.String, // optional graphql.NewNonNull(graphql.String),
					},
					"kind": &graphql.ArgumentConfig{
						Description: "kind of the linked objects",
						Type:        graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var group, kind string
					if v, ok := p.Args["group"]; ok {
						group = v.(string)
					}
					if v, ok := p.Args["kind"]; ok {
						kind = v.(string)
					}
					if group != "" && kind == "" { // group can be empty
						return nil, fmt.Errorf("group is set but kind is not set")
					}

					if oid, ok := p.Source.(*apiv1.ObjectID); ok {
						links, err := objGraph.Links(oid, edgeLabel)
						if err != nil {
							return nil, err
						}
						if kind != "" { // group can be empty
							linksForGK := links[metav1.GroupKind{Group: group, Kind: kind}]
							return linksForGK, nil
						}

						var out []apiv1.ObjectID
						for _, refs := range links {
							out = append(out, refs...)
						}
						return out, nil
					}
					return []interface{}{}, nil
				},
			})
		}(label)
	}

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"find": &graphql.Field{
				Type: oidType,
				Args: graphql.FieldConfigArgument{
					"key": &graphql.ArgumentConfig{
						Description: "Key of an object",
						Type:        graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					key := p.Args["key"].(string)
					return apiv1.ParseObjectID(apiv1.OID(key))
				},
			},
		},
	})
	StarWarsSchema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})

	h := handler.New(&handler.Config{
		Schema:     &StarWarsSchema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	return &StarWarsSchema, h
	//http.Handle("/", h)
	//log.Println("server running on port :8080")
	//http.ListenAndServe(":8080", nil)
}
