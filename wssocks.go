package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	armonsocks "github.com/armon/go-socks5"
	"github.com/kpango/glg"
	thingsocks "github.com/things-go/go-socks5"
	wzshisocks "github.com/wzshiming/socks5"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

var armonsocksServer *armonsocks.Server
var thingsocksServer *thingsocks.Server
var wzshisocksServer *wzshisocks.Server

func ws2socks(ws *websocket.Conn) {
	glg.Printf("receive ws: %p", ws)
	defer func() {
		_ = ws.Close()
		glg.Printf("finish ws: %p", ws)
	}()
	wzshisocksServer.ServeConn(ws)
}

func StartWsSocksServer(port int64) {
	cred := armonsocks.StaticCredentials{
		"yechenvk": "yechen123321",
	}
	cator := armonsocks.UserPassAuthenticator{Credentials: cred}
	armonsocksServer, _ = armonsocks.New(&armonsocks.Config{
		AuthMethods: []armonsocks.Authenticator{cator},
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 10*time.Second)
		},
	})

	cred1 := thingsocks.StaticCredentials{
		"yechenvk": "yechen123321",
	}
	cator1 := &thingsocks.UserPassAuthenticator{Credentials: cred1}
	origin, _ := url.Parse("/")
	thingsocksServer = thingsocks.NewServer(thingsocks.WithAuthMethods([]thingsocks.Authenticator{cator1}))

	logger := log.New(os.Stderr, "[socks5] ", log.LstdFlags)
	wzshisocksServer = &wzshisocks.Server{
		Logger:         logger,
		Authentication: wzshisocks.UserAuth("yechenvk", "yechen123321"),
	}

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
