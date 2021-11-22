package main

import (
	//   "context"

	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/go-chi/chi"
	"google.golang.org/grpc"
)

type programsResource struct{}
type CancelFunc func()

type responseData struct {
	Status  bool   `json:"Status"`
	Code    int    `json:"Code"`
	Object  string `json:"Object"`
	Message string `json:"Message"`
}

func (rs programsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/getAll", rs.GetAll)
	r.Get("/getById", rs.GetById)
	r.Get("/create", rs.Create)
	r.Get("/execute", rs.Execute)

	return r
}

// Request Handler - GET /AllPrograms - Read a list of programs.
func (rs programsResource) GetAll(w http.ResponseWriter, r *http.Request) {
	dg, cancel := getDgraphClient()
	defer cancel()
	type Programs struct {
		Uid          string `json:"uid,omitempty"`
		Name         string `json:"name,omitempty"`
		Codetex      string `json:"codetext,omitempty"`
		Language     string `json:"language,omitempty"`
		User         string `json:"user,omitempty"`
		CodeCompiled string `json:"codecompiled,omitempty"`
	}
	ctx := context.Background()

	variables := make(map[string]string)
	q := `
	{
		getAllPrograms(func: has(nameProgram)) {
		  nameProgram
		  user
		  language
		  codeTex
		  codeCompiled
		}
	  }
	   `
	resp, err := dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
	}
	type Root struct {
		Me []Programs `json:"me"`
	}
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resp.Json))
}
func (rs programsResource) GetById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Obtener por ID progra,"))
}
func (rs programsResource) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("CREATE progra,"))
}

func (rs programsResource) Execute(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("python3", "./assets/game.py")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	go copyOutput(stdout)
	go copyOutput(stderr)
	cmd.Wait()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Excute program"))
}

//ejecutar programs
func copyOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

// COnexion Grapgh QL
func getDgraphClient() (*dgo.Dgraph, CancelFunc) {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal("While trying to dial gRPC")
	}
	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)
	log.Printf("CONEXION EXITOSA CON DGRAPHQL")
	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error while closing connection:%v", err)
		}
	}
}
func getAllPrograms() {
	dg, cancel := getDgraphClient()
	defer cancel()
	type Programs struct {
		Uid          string `json:"uid,omitempty"`
		Name         string `json:"name,omitempty"`
		Codetex      string `json:"codetext,omitempty"`
		Language     string `json:"language,omitempty"`
		User         string `json:"user,omitempty"`
		CodeCompiled string `json:"codecompiled,omitempty"`
	}
	ctx := context.Background()

	variables := make(map[string]string)
	q := `
	{
		getAllPrograms(func: has(nameProgram)) {
		  nameProgram
		  user
		  language
		  codeTex
		  codeCompiled
		}
	  }
	   `
	resp, err := dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
	}
	type Root struct {
		Me []Programs `json:"me"`
	}
	var r Root
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(resp.Json))

}
