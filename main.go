package main

import (
	"log"
	"net/http"
)

func main() {
	//to run "go-bindata-assetfs static/..." in terminal
	http.Handle("/", http.FileServer(assetFS()))
	http.HandleFunc("/index", Index)
	http.HandleFunc("/list", List)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("start server failed, reson is : ", err)
	}
}
