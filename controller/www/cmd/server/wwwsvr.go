package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Commands returned from web server
const (
	CmdRelayOff = iota + 1
	CmdRelayOn
	CmdSetRelay
	CmdRelaySetPower
	CmdGetSensorValue
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// setActor handles route /setactor/{name}/{cmd}
func setActor(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	fmt.Println("setActor()")
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "%s=%s", vars["name"], vars["cmd"])

	svrChanOut <- ServerCommand{Cmd: CmdSetRelay, DeviceName: vars["name"], Value: []byte(vars["cmd"])}
}

func getSensorValue(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	//fmt.Println("getSensorValue()")
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	ret := make(chan string)
	svrChanOut <- ServerCommand{Cmd: CmdGetSensorValue, DeviceName: vars["name"], ChanReturn: ret}
	retValue := <-ret

	fmt.Println("getSensorValue return received: ", retValue)
	fmt.Fprintf(w, "%s", retValue)
}

type ServerCommand struct {
	Cmd        int
	DeviceName string
	Value      []byte
	ChanReturn chan string
}

type SvrChanIn chan ServerCommand
type SvrChanOut chan ServerCommand

var svrChanIn SvrChanIn
var svrChanOut SvrChanOut

func RunWebServer(in SvrChanIn, out SvrChanOut) {

	svrChanIn = in
	svrChanOut = out

	r := mux.NewRouter()

	r.HandleFunc("/setactor/{name}/{cmd}", setActor)
	r.HandleFunc("/getsensor/{name}", getSensorValue)

	// This will serve files under http://localhost:8000/static/<filename>
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("www/assets"))))

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
