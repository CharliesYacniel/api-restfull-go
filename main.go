package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"google.golang.org/grpc"
)

type Link struct {
	Uid   string   `json:"uid,omitempty"`
	URL   string   `json:"url,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}
type CancelFunc func()

func main() {
	port := "3000"

	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	log.Printf("API KENIA en LINEA = http://localhost:%s", port)

	//CONEXION BASE DE DATOS
	getDgraphClient()

	//CONSULTAS A BD
	ExampleTxn_Mutate()

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API HOME!"))
	})

	r.Mount("/posts", postsResource{}.Routes())
	r.Mount("/programs", programsResource{}.Routes())

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getDgraphClient() (*dgo.Dgraph, CancelFunc) {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal("While trying to dial gRPC")
	}

	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)
	log.Printf("CONEXION EXITOSA CON DGRAPHQL")
	// ctx := context.Background()

	// Perform login call. If the Dgraph cluster does not have ACL and
	// enterprise features enabled, this call should be skipped.
	// for {
	// 	// Keep retrying until we succeed or receive a non-retriable error.
	// 	// err = dg.Login(ctx, "groot", "password")
	// 	// err = dg.Login(ctx, "groot", "password")
	// 	// err = dg.Login(ctx)
	// 	// fmt.Println(err)
	// 	if err == nil || !strings.Contains(err.Error(), "Please retry") {
	// 		break
	// 	}
	// 	time.Sleep(time.Second)
	// }
	// if err != nil {
	// 	log.Fatalf("While trying to login %v", err.Error())
	// }

	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error while closing connection:%v", err)
		}

	}
}

func ExampleTxn_Mutate() {
	type School struct {
		Name  string   `json:"name,omitempty"`
		DType []string `json:"dgraph.type,omitempty"`
	}

	type loc struct {
		Type   string    `json:"type,omitempty"`
		Coords []float64 `json:"coordinates,omitempty"`
	}

	// If omitempty is not set, then edges with empty values (0 for int/float, "" for string, false
	// for bool) would be created for values not specified explicitly.

	type Person struct {
		Uid      string   `json:"uid,omitempty"`
		Name     string   `json:"name,omitempty"`
		Age      int      `json:"age,omitempty"`
		Married  bool     `json:"married,omitempty"`
		Raw      []byte   `json:"raw_bytes,omitempty"`
		Friends  []Person `json:"friends,omitempty"`
		Location loc      `json:"loc,omitempty"`
		School   []School `json:"school,omitempty"`
		DType    []string `json:"dgraph.type,omitempty"`
	}

	dg, cancel := getDgraphClient()
	defer cancel()
	// While setting an object if a struct has a Uid then its properties in the
	// graph are updated else a new node is created.
	// In the example below new nodes for Alice, Bob and Charlie and school
	// are created (since they don't have a Uid).
	p := Person{
		Uid:     "_:alice",
		Name:    "Alice",
		Age:     26,
		Married: true,
		DType:   []string{"Person"},
		Location: loc{
			Type:   "Point",
			Coords: []float64{1.1, 2},
		},
		Raw: []byte("raw_bytes"),
		Friends: []Person{{
			Name:  "Bob",
			Age:   24,
			DType: []string{"Person"},
		}, {
			Name:  "Charlie",
			Age:   29,
			DType: []string{"Person"},
		}},
		School: []School{{
			Name:  "Crown Public School",
			DType: []string{"Institution"},
		}},
	}

	op := &api.Operation{}
	op.Schema = `
		   name: string @index(exact) .
		   age: int .
		   married: bool .
		   Friends: [uid] .
		   loc: geo .
		   type: string .
		   coords: float .
		   Friends: [uid] .
   
		   type Person {
			   name
			   age
			   married
			   Friends
			   loc
			 }
   
		   type Institution {
			   name
		   }
	   `

	ctx := context.Background()
	if err := dg.Alter(ctx, op); err != nil {
		log.Fatal(err)
	}

	mu := &api.Mutation{
		CommitNow: true,
	}
	pb, err := json.Marshal(p)
	if err != nil {
		log.Fatal(err)
	}

	mu.SetJson = pb
	response, err := dg.NewTxn().Mutate(ctx, mu)
	if err != nil {
		log.Fatal(err)
	}

	// Assigned uids for nodes which were created would be returned in the response.Uids map.
	puid := response.Uids["alice"]
	const q = `
		   query Me($id: string){
			   me(func: uid($id)) {
				   name
				   age
				   loc
				   raw_bytes
				   married
				   dgraph.type
				   friends @filter(eq(name, "Bob")) {
					   name
					   age
					   dgraph.type
				   }
				   school {
					   name
					   dgraph.type
				   }
			   }
		   }
	   `

	variables := make(map[string]string)
	variables["$id"] = puid
	resp, err := dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Me []Person `json:"me"`
	}

	var r Root
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(r, "", "\t")
	fmt.Printf("%s\n", out)
	// Output: {
	// 	"me": [
	// 		{
	// 			"name": "Alice",
	// 			"age": 26,
	// 			"married": true,
	// 			"raw_bytes": "cmF3X2J5dGVz",
	// 			"friends": [
	// 				{
	// 					"name": "Bob",
	// 					"age": 24,
	// 					"loc": {},
	// 					"dgraph.type": [
	// 						"Person"
	// 					]
	// 				}
	// 			],
	// 			"loc": {
	// 				"type": "Point",
	// 				"coordinates": [
	// 					1.1,
	// 					2
	// 				]
	// 			},
	// 			"school": [
	// 				{
	// 					"name": "Crown Public School",
	// 					"dgraph.type": [
	// 						"Institution"
	// 					]
	// 				}
	// 			],
	// 			"dgraph.type": [
	// 				"Person"
	// 			]
	// 		}
	// 	]
	// }

}
