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
	CodeTex      string `json:"codeTex,omitempty"`
	Language     string `json:"language,omitempty"`
	User         string `json:"user,omitempty"`
	CodeCompiled string `json:"codeCompiled,omitempty"`
}

type responseData struct {
	Status  bool   `json:"status"`
	Code    int    `json:"code"`
	Obj     string `json:"obj"`
	Message string `json:"message"`
}

func (rs programsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/getAll", rs.GetAll)
	r.Get("/getById", rs.GetById)
	r.Post("/create", rs.Create)
	r.Put("/update", rs.Update)
	r.Post("/execute", rs.Execute)

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

	r.ParseForm()
	nameProgram := r.Form.Get("nameProgram")
	user := r.Form.Get("user")
	language := r.Form.Get("language")
	codeTex := r.Form.Get("codeTex")
	codeCompiled := r.Form.Get("codeCompiled")
	if nameProgram == "" || user == "" || language == "" || codeTex == "" || codeCompiled == "" {
		log.Fatal("Search query not found!")
		return
	}
	p := Programs{
		Uid:          "_:prog",
		NameProgram:  nameProgram,
		CodeTex:      codeTex,
		Language:     language,
		User:         user,
		CodeCompiled: codeCompiled,
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
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resp.Json))
}

func (rs programsResource) Update(w http.ResponseWriter, r *http.Request) {
	dg, cancel := getDgraphClient()
	defer cancel()
	ctx := context.Background()

	r.ParseForm()

	uid := r.Form.Get("uid")
	nameProgram := r.Form.Get("nameProgram")
	user := r.Form.Get("user")
	language := r.Form.Get("language")
	codeTex := r.Form.Get("codeTex")
	codeCompiled := r.Form.Get("codeCompiled")
	if nameProgram == "" || user == "" || language == "" || codeTex == "" || codeCompiled == "" {
		log.Fatal("Search query not found!")
		return
	}

	query := `
		query {
			prog as var(func: uid(` + uid + `))
		}`
	mu := &api.Mutation{
		SetNquads: []byte(` uid(prog)  <codeTex> "` + codeTex + `" . `),
	}
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}
	// Update email only if matching uid found.
	resp, err := dg.NewTxn().Do(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resp.Json))
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
	result, _ := json.Marshal(responseData{
		true,
		200,
		"objeto compilado",
		"Compile",
	})
	io.WriteString(w, string(result))
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
