package main

import (
	"PaoDeKuaiServer/socket"
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8181", "http service address")

func main() {
	flag.Parse()
	hub := socket.NewHub()
	go hub.Run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}