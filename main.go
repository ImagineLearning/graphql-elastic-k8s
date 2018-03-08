package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/graphql-go/graphql"

	"github.com/olivere/elastic"
)

var (
	Luke           StarWarsChar
	Vader          StarWarsChar
	Han            StarWarsChar
	Leia           StarWarsChar
	Tarkin         StarWarsChar
	Threepio       StarWarsChar
	Artoo          StarWarsChar
	HumanData      map[int]StarWarsChar
	DroidData      map[int]StarWarsChar
	StarWarsSchema graphql.Schema

	humanType *graphql.Object
	droidType *graphql.Object
	client    *elastic.Client
)

type StarWarsChar struct {
	ID              string
	Name            string
	Friends         []StarWarsChar
	AppearsIn       []int
	HomePlanet      string
	PrimaryFunction string
}

func init() {
	var err error
	client, err = elastic.NewClient(elastic.SetURL("http://elasticsearch.elasticsearch.svc.cluster.local:9200"))
	if err != nil {
		fmt.Print("Error connecting to elasticsearch: " + err.Error())
		return
	}

	Luke = StarWarsChar{
		ID:         "1000",
		Name:       "Luke Skywalker",
		AppearsIn:  []int{4, 5, 6},
		HomePlanet: "Tatooine",
	}
	Vader = StarWarsChar{
		ID:         "1001",
		Name:       "Darth Vader",
		AppearsIn:  []int{4, 5, 6},
		HomePlanet: "Tatooine",
	}
	Han = StarWarsChar{
		ID:        "1002",
		Name:      "Han Solo",
		AppearsIn: []int{4, 5, 6},
	}
	Leia = StarWarsChar{
		ID:         "1003",
		Name:       "Leia Organa",
		AppearsIn:  []int{4, 5, 6},
		HomePlanet: "Alderaa",
	}
	Tarkin = StarWarsChar{
		ID:        "1004",
		Name:      "Wilhuff Tarkin",
		AppearsIn: []int{4},
	}
	Threepio = StarWarsChar{
		ID:              "2000",
		Name:            "C-3PO",
		AppearsIn:       []int{4, 5, 6},
		PrimaryFunction: "Protocol",
	}
	Artoo = StarWarsChar{
		ID:              "2001",
		Name:            "R2-D2",
		AppearsIn:       []int{4, 5, 6},
		PrimaryFunction: "Astromech",
	}
	Luke.Friends = append(Luke.Friends, []StarWarsChar{Han, Leia, Threepio, Artoo}...)
	Vader.Friends = append(Luke.Friends, []StarWarsChar{Tarkin}...)
	Han.Friends = append(Han.Friends, []StarWarsChar{Luke, Leia, Artoo}...)
	Leia.Friends = append(Leia.Friends, []StarWarsChar{Luke, Han, Threepio, Artoo}...)
	Tarkin.Friends = append(Tarkin.Friends, []StarWarsChar{Vader}...)
	Threepio.Friends = append(Threepio.Friends, []StarWarsChar{Luke, Han, Leia, Artoo}...)
	Artoo.Friends = append(Artoo.Friends, []StarWarsChar{Luke, Han, Leia}...)
	HumanData = map[int]StarWarsChar{
		1000: Luke,
		1001: Vader,
		1002: Han,
		1003: Leia,
		1004: Tarkin,
	}
	DroidData = map[int]StarWarsChar{
		2000: Threepio,
		2001: Artoo,
	}

	episodeEnum := graphql.NewEnum(graphql.EnumConfig{
		Name:        "Episode",
		Description: "One of the films in the Star Wars Trilogy",
		Values: graphql.EnumValueConfigMap{
			"NEWHOPE": &graphql.EnumValueConfig{
				Value:       4,
				Description: "Released in 1977.",
			},
			"EMPIRE": &graphql.EnumValueConfig{
				Value:       5,
				Description: "Released in 1980.",
			},
			"JEDI": &graphql.EnumValueConfig{
				Value:       6,
				Description: "Released in 1983.",
			},
		},
	})

	characterInterface := graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "Character",
		Description: "A character in the Star Wars Trilogy",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The id of the character.",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The name of the character.",
			},
			"appearsIn": &graphql.Field{
				Type:        graphql.NewList(episodeEnum),
				Description: "Which movies they appear in.",
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			if character, ok := p.Value.(StarWarsChar); ok {
				id, _ := strconv.Atoi(character.ID)
				human := GetHuman(id)
				if human.ID != "" {
					return humanType
				}
			}
			return droidType
		},
	})
	characterInterface.AddFieldConfig("friends", &graphql.Field{
		Type:        graphql.NewList(characterInterface),
		Description: "The friends of the character, or an empty list if they have none.",
	})

	humanType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Human",
		Description: "A humanoid creature in the Star Wars universe.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The id of the human.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if human, ok := p.Source.(StarWarsChar); ok {
						return human.ID, nil
					}
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The name of the human.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if human, ok := p.Source.(StarWarsChar); ok {
						return human.Name, nil
					}
					return nil, nil
				},
			},
			"friends": &graphql.Field{
				Type:        graphql.NewList(characterInterface),
				Description: "The friends of the human, or an empty list if they have none.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if human, ok := p.Source.(StarWarsChar); ok {
						return human.Friends, nil
					}
					return []interface{}{}, nil
				},
			},
			"appearsIn": &graphql.Field{
				Type:        graphql.NewList(episodeEnum),
				Description: "Which movies they appear in.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if human, ok := p.Source.(StarWarsChar); ok {
						return human.AppearsIn, nil
					}
					return nil, nil
				},
			},
			"homePlanet": &graphql.Field{
				Type:        graphql.String,
				Description: "The home planet of the human, or null if unknown.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if human, ok := p.Source.(StarWarsChar); ok {
						return human.HomePlanet, nil
					}
					return nil, nil
				},
			},
		},
		Interfaces: []*graphql.Interface{
			characterInterface,
		},
	})
	droidType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Droid",
		Description: "A mechanical creature in the Star Wars universe.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The id of the droid.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if droid, ok := p.Source.(StarWarsChar); ok {
						return droid.ID, nil
					}
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The name of the droid.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if droid, ok := p.Source.(StarWarsChar); ok {
						return droid.Name, nil
					}
					return nil, nil
				},
			},
			"friends": &graphql.Field{
				Type:        graphql.NewList(characterInterface),
				Description: "The friends of the droid, or an empty list if they have none.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if droid, ok := p.Source.(StarWarsChar); ok {
						friends := []map[string]interface{}{}
						for _, friend := range droid.Friends {
							friends = append(friends, map[string]interface{}{
								"name": friend.Name,
								"id":   friend.ID,
							})
						}
						return droid.Friends, nil
					}
					return []interface{}{}, nil
				},
			},
			"appearsIn": &graphql.Field{
				Type:        graphql.NewList(episodeEnum),
				Description: "Which movies they appear in.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if droid, ok := p.Source.(StarWarsChar); ok {
						return droid.AppearsIn, nil
					}
					return nil, nil
				},
			},
			"primaryFunction": &graphql.Field{
				Type:        graphql.String,
				Description: "The primary function of the droid.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if droid, ok := p.Source.(StarWarsChar); ok {
						return droid.PrimaryFunction, nil
					}
					return nil, nil
				},
			},
		},
		Interfaces: []*graphql.Interface{
			characterInterface,
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"hero": &graphql.Field{
				Type: characterInterface,
				Args: graphql.FieldConfigArgument{
					"episode": &graphql.ArgumentConfig{
						Description: "If omitted, returns the hero of the whole saga. If " +
							"provided, returns the hero of that particular episode.",
						Type: episodeEnum,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetHero(p.Args["episode"]), nil
				},
			},
			"human": &graphql.Field{
				Type: humanType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Description: "id of the human",
						Type:        graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, err := strconv.Atoi(p.Args["id"].(string))
					if err != nil {
						return nil, err
					}
					return GetHuman(id), nil
				},
			},
			"droid": &graphql.Field{
				Type: droidType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Description: "id of the droid",
						Type:        graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, err := strconv.Atoi(p.Args["id"].(string))
					if err != nil {
						return nil, err
					}
					return GetDroid(id), nil
				},
			},
		},
	})
	StarWarsSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
}

func GetHuman(id int) StarWarsChar {
	// Search with a term query
	termQuery := elastic.NewTermQuery("ID", id)
	searchResult, err := client.Search().
		Index("starwars").       // search in index "tweets"
		Query(termQuery).        // specify the query
		From(0).Size(1).         // take document 0
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		// Handle error
		fmt.Print("Error searching elasticsearch: " + err.Error())
		return StarWarsChar{}
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits.TotalHits > 0 {
		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {

			// Deserialize hit.Source into a StarWarsChar
			var t StarWarsChar
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				// Deserialization failed
				fmt.Printf("Failed to deserialize StarWars: " + err.Error())
				return StarWarsChar{}
			}

			return t
		}
	} else {
		// No hits
		fmt.Print("Found no hits\n")
		return StarWarsChar{}
	}

	return StarWarsChar{}
}
func GetDroid(id int) StarWarsChar {
	if droid, ok := DroidData[id]; ok {
		return droid
	}
	return StarWarsChar{}
}
func GetHero(episode interface{}) interface{} {
	if episode == 5 {
		return Luke
	}
	return Artoo
}

func main() {

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")

		result := graphql.Do(graphql.Params{
			Schema:        StarWarsSchema,
			RequestString: query,
		})
		json.NewEncoder(w).Encode(result)
	})
	fmt.Println("Now server is running on port 6000")
	http.ListenAndServe(":6000", nil)
}
