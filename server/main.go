package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
)

var addr = flag.String("addr", GetOutboundIP().String()+":8285", "http service address")

func main() {
	flag.Parse()
	go world.Run()

	fmt.Println(*addr)
	log.Fatal(http.ListenAndServe(*addr, httpHandler()))
}

func httpHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/connect", connect)
	return mux
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
