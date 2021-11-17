package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Link struct {
	Uid   string   `json:"uid,omitempty"`
	URL   string   `json:"url,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

func main() {
	port := "8080"

	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	log.Printf("API KENIA en LINEA = http://localhost:%s", port)

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API HOME!"))
	})

	r.Mount("/posts", postsResource{}.Routes())
	r.Mount("/programs", programsResource{}.Routes())

	log.Fatal(http.ListenAndServe(":"+port, r))

	// ctx := context.TODO()

	//    conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	//    if err != nil {
	//       log.Fatal("failed to dial ", err)
	//    }
	//    defer conn.Close()

	//    dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	//    txn := dgraphClient.NewTxn()
	//    defer txn.Commit(ctx)

	//    url := fmt.Sprintf("https://example.com/%v", time.Now().UnixNano())

	//    link := Link{
	//       URL:   url,
	//       DType: []string{"Link"},
	//    }

	//    lb, err := json.Marshal(link)
	//    if err != nil {
	//       log.Fatal("failed to marshal ", err)
	//    }

	//    mu := &api.Mutation{
	//       SetJson: lb,
	//    }
	//    res, err := txn.Mutate(ctx, mu)
	//    if err != nil {
	//       log.Fatal("failed to mutate ", err)
	//    }

	//    print("res: %v", res)
}
