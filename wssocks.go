package main

import (
	"fmt"
	"net/http"

	"github.com/armon/go-socks5"
	"github.com/kpango/glg"
	"golang.org/x/net/websocket"
)

var socksServer *socks5.Server

func ws2socks(ws *websocket.Conn) {
	glg.Printf("receive ws: %p", ws)
	defer func() {
		glg.Printf("finish ws: %p", ws)
	}()
	err := socksServer.ServeConn(ws)
	if err != nil {
		glg.Errorf("ws serve error:", err)
		_ = ws.Close()
		return
	}
}

func StartWsSocksServer(port int64) {
	cred := socks5.StaticCredentials{
		"yechenvk": "yechen123321",
	}
	cator := socks5.UserPassAuthenticator{Credentials: cred}
	socksServer, _ = socks5.New(&socks5.Config{
		AuthMethods: []socks5.Authenticator{cator},
	})
	http.Handle("/wssocks", &websocket.Server{
		Handler: ws2socks,
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	})
	glg.Infof("StartWsSocksServer=%v", port)
	var runPort = fmt.Sprintf(":%v", port)
	glg.Fatalln(http.ListenAndServe(runPort, nil))
}