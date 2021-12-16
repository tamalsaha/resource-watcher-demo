package main

import (
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/examples/todo/schema"
	"github.com/graphql-go/graphql/testutil"
	"github.com/graphql-go/handler"
	apiv1 "kmodules.xyz/client-go/api/v1"
	"log"
	"net/http"
)

// https://github.com/graphql-go/graphql/blob/master/examples/star-wars/main.go
func main() {
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
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if obj, ok := p.Source.(apiv1.ObjectID); ok {
						return obj.Group, nil
					}
					return nil, nil
				},
			},
			"kind": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The kind of the Object",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if obj, ok := p.Source.(apiv1.ObjectID); ok {
						return obj.Kind, nil
					}
					return nil, nil
				},
			},
			"namespace": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The namespace of the Object",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if obj, ok := p.Source.(apiv1.ObjectID); ok {
						return obj.Namespace, nil
					}
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The name of the human.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if obj, ok := p.Source.(apiv1.ObjectID); ok {
						return obj.Name, nil
					}
					return nil, nil
				},
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
				Type:        graphql.NewNonNull(graphql.String),
			},
			"kind": &graphql.ArgumentConfig{
				Description: "kind of the linked objects",
				Type:        graphql.NewNonNull(graphql.String),
			},
		},
		//Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		//	group, err := strconv.Atoi(p.Args["group"].(string))
		//	if err != nil {
		//		return nil, err
		//	}
		//	kind, err := strconv.Atoi(p.Args["kind"].(string))
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	if obj, ok := p.Source.(apiv1.ObjectID); ok {
		//		return obj.Friends, nil
		//	}
		//	return []interface{}{}, nil
		//},
	})

	h := handler.New(&handler.Config{
		Schema:     &testutil.StarWarsSchema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	http.Handle("/", h)
	log.Println("server running on port :8080")
	http.ListenAndServe(":8080", nil)
}

// go run gql/*.go
// https://wehavefaces.net/learn-golang-graphql-relay-1-e59ea174a902
// https://github.com/graphql-go/graphql/pull/574
func main__() {
	h := handler.New(&handler.Config{
		Schema:     &schema.TodoSchema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	http.Handle("/", h)
	log.Println("server running on port :8080")
	http.ListenAndServe(":8080", nil)
}

func main_() {
	// Schema
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	// Query
	query := `
		{
			hello
		}
	`
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("%s \n", rJSON) // {"data":{"hello":"world"}}
}
