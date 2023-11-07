package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/armon/go-socks5"
	"github.com/kpango/glg"
	gosocks5 "github.com/things-go/go-socks5"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

var socksServer *socks5.Server
var socksServer2 *gosocks5.Server

func ws2socks(ws *websocket.Conn) {
	glg.Printf("receive ws: %p", ws)
	defer func() {
		_ = ws.Close()
		glg.Printf("finish ws: %p", ws)
	}()
	socksServer2.ServeConn(ws)
	err := socksServer.ServeConn(ws)
	if err != nil {
		glg.Errorf("ws serve error %v:", err)
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
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 10*time.Second)
		},
	})
	cred1 := gosocks5.StaticCredentials{
		"yechenvk": "yechen123321",
	}
	cator1 := &gosocks5.UserPassAuthenticator{Credentials: cred1}
	origin, _ := url.Parse("/")
	socksServer2 = gosocks5.NewServer(gosocks5.WithAuthMethods([]gosocks5.Authenticator{cator1}))
	http.Handle("/wssocks", &websocket.Server{Handler: ws2socks, Config: websocket.Config{
		Origin: origin,
	}})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	})
	glg.Infof("StartWsSocksServer=%v", port)
	var runPort = fmt.Sprintf(":%v", port)
	glg.Fatalln(http.ListenAndServe(runPort, nil))
}
