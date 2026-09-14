package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/m/", handPage)

	log.Println("listen ...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
