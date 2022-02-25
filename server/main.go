package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	http.HandleFunc("/connect", connect)
	http.HandleFunc("/chat", chatLoop)
	http.HandleFunc("/game", gameLoop)

	fmt.Printf(*addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
