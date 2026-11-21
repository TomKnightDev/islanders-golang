package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
)

var (
	addr = flag.String("addr", GetOutboundIP().String()+":8285", "http service address")
	dev  = flag.Bool("dev", false, "allow dev tools, such as changing the game speed, for testing")
)

func main() {
	flag.Parse()
	world.dev = *dev
	if *dev {
		log.Println("Dev tools are on")
	}
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
