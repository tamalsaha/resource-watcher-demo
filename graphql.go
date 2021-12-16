package main

import (
	"fmt"
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
				//Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				//	if obj, ok := p.Source.(apiv1.ObjectID); ok {
				//		return obj.Group, nil
				//	}
				//	return nil, nil
				//},
			},
			"kind": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The kind of the Object",
				//Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				//	if obj, ok := p.Source.(apiv1.ObjectID); ok {
				//		return obj.Kind, nil
				//	}
				//	return nil, nil
				//},
			},
			"namespace": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The namespace of the Object",
				//Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				//	if obj, ok := p.Source.(apiv1.ObjectID); ok {
				//		return obj.Namespace, nil
				//	}
				//	return nil, nil
				//},
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The name of the human.",
				//Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				//	if obj, ok := p.Source.(apiv1.ObjectID); ok {
				//		return obj.Name, nil
				//	}
				//	return nil, nil
				//},
			},
			//"friends": &graphql.Field{
			//	Type:        graphql.NewList(characterInterface),
			//	Description: "The friends of the human, or an empty list if they have none.",
			//	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			//		if human, ok := p.Source.(StarWarsChar); ok {
			//			return human.Friends, nil
			//		}
			//		return []interface{}{}, nil
			//	},
			//},
			//"appearsIn": &graphql.Field{
			//	Type:        graphql.NewList(episodeEnum),
			//	Description: "Which movies they appear in.",
			//	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			//		if human, ok := p.Source.(StarWarsChar); ok {
			//			return human.AppearsIn, nil
			//		}
			//		return nil, nil
			//	},
			//},
			//"homePlanet": &graphql.Field{
			//	Type:        graphql.String,
			//	Description: "The home planet of the human, or null if unknown.",
			//	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			//		if human, ok := p.Source.(StarWarsChar); ok {
			//			return human.HomePlanet, nil
			//		}
			//		return nil, nil
			//	},
			//},
		},
	})
	oidType.AddFieldConfig("links", &graphql.Field{
		Type:        graphql.NewList(oidType),
		Description: "Links from this object",
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
			if group == "" && kind != "" {
				return nil, fmt.Errorf("group is not set but kind is set")
			} else if group != "" && kind == "" {
				return nil, fmt.Errorf("group is set but kind is not set")
			}

			if oid, ok := p.Source.(*apiv1.ObjectID); ok {
				links, err := objGraph.Links(oid)
				if err != nil {
					return nil, err
				}
				if group != "" && kind != "" {
					return links[metav1.GroupKind{Group: group, Kind: kind}], nil
				}

				var out []apiv1.ObjectID
				for gk, refs := range links {
					for _, ref := range refs {
						out = append(out, apiv1.ObjectID{
							Group:     gk.Group,
							Kind:      gk.Kind,
							Namespace: ref.Namespace,
							Name:      ref.Name,
						})
					}
				}
				return out, nil
			}
			return []interface{}{}, nil
		},
	})

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
					return apiv1.ParseObjectID(key)
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
