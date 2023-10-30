package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "wss://"+r.Host+"/echo")
}

func StartWebSocketServer() {
	var port = "8080"
	if name := os.Getenv("PORT_ENV_NAME"); name != "" {
		port = os.Getenv(name)
		log.Printf("Get PORT_ENV_NAME=%v, try get env port=%v", name, port)
	}
	log.Printf("final use port=%v", port)
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))

func IpPacketServer() {
	netaddr, err := net.ResolveIPAddr("ip4", "10.82.170.194")
	if err != nil {
		fmt.Println("Server ResolveIPAddr err:", err)
		return
	}
	conn, err := net.ListenIP("ip4:tcp", netaddr)
	if err != nil {
		fmt.Println("Server DialIP failed:", err)
		return
	}
	fmt.Println("Server Listen success")
	for {
		data := make([]byte, 128)
		readLen, remoteAddr, err := conn.ReadFromIP(data)
		if err != nil {
			fmt.Println("Server ReadFromIP err:", err)
			return
		}
		fmt.Println("Server readFromIp:", string(data[:readLen]))
		_, err = conn.WriteToIP([]byte("哈哈你好啊"), remoteAddr)
		if err != nil {
			fmt.Println("Server WriteToIP err:", err)
			return
		}
	}
}

func IpPacketClient() {
	netaddr, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		fmt.Println("Client ResolveIPAddr err:", err)
		return
	}
	conn, err := net.ListenIP("ip4:tcp", netaddr)
	if err != nil {
		fmt.Println("Client DialIP failed:", err)
		return
	}
	fmt.Println("Client Listen success")
	remoteAddr, err := net.ResolveIPAddr("ip4:tcp", "10.82.170.194")
	var count = 0
	for {
		time.Sleep(20 * time.Millisecond)
		if count > 10 {
			break
		}
		count++
		_, err = conn.WriteToIP([]byte("哈哈你好啊"), remoteAddr)
		if err != nil {
			fmt.Println("Client WriteToIP err:", err)
			return
		}
		data := make([]byte, 128)
		readLen, _, err := conn.ReadFromIP(data)
		if err != nil {
			fmt.Println("Client ReadFromIP err:", err)
			return
		}
		fmt.Println("Client readFromIp:", string(data[:readLen]))
	}
}

func main() {
	go IpPacketServer()
	time.Sleep(1 * time.Second)
	IpPacketClient()
	fmt.Println("进程退出")
}
