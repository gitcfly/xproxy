package main

import (
	"io"
	"net"
	"runtime/debug"

	"github.com/kpango/glg"
)

var tcbMap = make(map[string]*TCB)

func StartTcpProxy() {
	glg.Get().SetMode(glg.BOTH).SetLineTraceMode(glg.TraceLineShort)
	ser, err := net.Listen("tcp", ":8080")
	if err != nil {
		glg.Errorf("proxyServer Listen err=%v", err)
		return
	}
	for {
		clientConn, err := ser.Accept()
		if err != nil {
			glg.Errorf("proxyServer accept err=%v", err)
			continue
		}
		go HandleClientTcpConnection(clientConn)
	}
}

func HandleClientTcpConnection(clientConn net.Conn) {
	defer HandlePanicError()
	for {
		buf, err := ReadPacket(clientConn)
		if err != nil && err != io.EOF {
			glg.Errorf("读取客户端数据失败 err=%v", err)
			return
		}
		if len(buf) == 0 {
			glg.Errorf("读取客户端数据失败 长度为0", err)
			continue
		}
		if string(buf) == "abcd" {
			continue
		}
		go HandleClientInputData(clientConn, buf)
	}
}

func HandleClientInputData(clientConn net.Conn, data []byte) {
	defer HandlePanicError()
	var buffer = BytesBufferWrap(data)
	var packet = NewPacket(buffer)
	switch {
	case packet.isTCP:
		glg.Successf("packet is tcp")
		ClientTcpInput(clientConn, packet)
	case packet.isUDP:
		glg.Successf("packet is udp")
		ClientUdpInput(clientConn, packet)
		return
	default:
		glg.Infof("packet is other=%v", int64(packet.ip4Header.protocol))
		return
	}
}

func HandlePanicError() {
	if err := recover(); err != nil {
		glg.Errorf("panic: %v", string(debug.Stack()))
	}
}

func sendDataToClient(conn net.Conn, buffer *BytesBuffer) error {
	_, err := WritePacket(conn, buffer.getFullBytes())
	if err != nil {
		glg.Errorf("sendDataToClient write data err=%v", err)
		return err
	}
	return nil
}

func sendDataToRemote(conn net.Conn, headerLen int32, buffer *BytesBuffer) error {
	data := buffer.getFullBytes()
	_, err := conn.Write(data[headerLen:])
	if err != nil {
		glg.Errorf("sendDataToRemote write data err=%v", err)
		return err
	}
	return nil
}
