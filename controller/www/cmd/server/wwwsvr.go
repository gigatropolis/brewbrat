package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// setActor handles route /setactor/{name}/{cmd}
func setActor(w http.ResponseWriter, r *http.Request) {
	fmt.Println("setActor()")
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", vars["cmd"])
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/setactor/{name}/{cmd}", setActor)

	// This will serve files under http://localhost:8000/static/<filename>
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("../../assets"))))

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8090",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

	/*err := http.ListenAndServe(":9090", http.FileServer(http.Dir("../../assets")))
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}*/
}
