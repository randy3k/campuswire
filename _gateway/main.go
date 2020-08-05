package main

import (
	"net/http"
	"log"
	"example.com/campuswire"
)


func main() {
	http.HandleFunc("/", campuswire.CampusWire)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
