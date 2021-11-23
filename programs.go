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
type Programs struct {
	Uid          string `json:"uid,omitempty"`
	NameProgram  string `json:"nameProgram,omitempty"`
	CodeTex      string `json:"codeText,omitempty"`
	Language     string `json:"language,omitempty"`
	User         string `json:"user,omitempty"`
	CodeCompiled string `json:"codeCompiled,omitempty"`
}

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
	r.Get("/update", rs.Update)
	r.Get("/execute", rs.Execute)

	return r
}

// Request Handler - GET /AllPrograms - Read a list of programs.
func (rs programsResource) GetAll(w http.ResponseWriter, r *http.Request) {
	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()

	variables := make(map[string]string)
	q := `
	{
		getAll(func: has(nameProgram)) {
		  uid
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

	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()

	variables := make(map[string]string)
	q := `
	{
		getByUid(func: uid("0x9c53")) {
		  uid
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

func (rs programsResource) Create(w http.ResponseWriter, r *http.Request) {

	dg, cancel := getDgraphClient()
	defer cancel()
	p := Programs{
		Uid:          "_:prog",
		NameProgram:  "Program two",
		CodeTex:      "import sys i=0 while i<10: print ('Hello') sys.stdout.flush() i=i+1 # time.sleep(1) n1 = 3 n2 = 10 suma = n1+n2 print('La suma es: ', suma)",
		Language:     "tasdasdrue",
		User:         "tasdasdrue",
		CodeCompiled: "tasdasdrue",
	}

	op := &api.Operation{}
	op.Schema = `
		nameProgram: string .
		user: string .
		language: string .
		codeTex: string .
		codeCompiled: string .
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
	variables := map[string]string{"$id1": response.Uids["prog"]}
	q := `query Create($id1: string){
		create(func: uid($id1)) {
			uid
			nameProgram
			user
			language
			codeTex
			codeCompiled
		}
	}`

	resp, err := dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Me []Programs `json:"me"`
	}

	// var r Root
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(r, "", "\t")
	fmt.Printf("%s\n", out)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resp.Json))
}

func (rs programsResource) Update(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	// w.Write([]byte(resp.Json))
	w.Write([]byte("uodate"))
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
