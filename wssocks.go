package main

import (
	"log"
	"net/http"

	"github.com/armon/go-socks5"
	"github.com/kpango/glg"
	"golang.org/x/net/websocket"
)

var socks, _ = socks5.New(&socks5.Config{})

func ws2socks(ws *websocket.Conn) {
	log.Printf("[INFO] receive ws: %p\n", ws)
	defer func() {
		log.Printf("[INFO] close ws: %p\n", ws)
		_ = ws.Close()
	}()
	err := socks.ServeConn(ws)
	if err != nil {
		glg.Errorf("[ERROR] ws serve error:", err)
		return
	}
}

func StartWsSocksServer() {
	http.Handle("/wssocks", websocket.Handler(ws2socks))
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	})
	glg.Fatalln(http.ListenAndServe(":8080", nil))
}
