package main

import (
	//   "context"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/go-chi/chi"
	// "github.com/go-chi/chi"
)

type programsResource struct{}

type responseData struct {
	Status  bool   `json:"Status"`
	Code    int    `json:"Code"`
	Object  string `json:"Object"`
	Message string `json:"Message"`
}

func (rs programsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/getAll", rs.List) // GET /programs - Read a list of programs.
	//   route.Post("/", rs.Create) // POST /programs - Create a new post.

	//   route.Route("/{id}", func(r chi.Router) {
	//     route.Use(PostCtx)
	//     route.Get("/", rs.Get)       // GET /programs/{id} - Read a single post by :id.
	//     route.Put("/", rs.Update)    // PUT /programs/{id} - Update a single post by :id.
	//     route.Delete("/", rs.Delete) // DELETE /programs/{id} - Delete a single post by :id.
	//   })

	return r
}

// Request Handler - GET /AllPrograms - Read a list of programs.
func (rs programsResource) List(w http.ResponseWriter, r *http.Request) {

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

	// res2D := &responseData{
	// 	Status:  true,
	// 	Code:    200,
	// 	Object:  `{"id": 1,"code": "n1 = float(input('Intro número uno: ')) n2 = float(input('Intro numero dos: ')) suma = n1+n2 print('La suma es: ', suma) n3 = float(input('la resuma:')) l = suma+n3 print('la resuma es:', l)"}`,
	// 	Message: `datos obtenidos con exito`,
	// }

	// res2B, _ := json.Marshal(res2D)
	// fmt.Println(string(res2B))

	//   resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	//   resp = res2D
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func copyOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
