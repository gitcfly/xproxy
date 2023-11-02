package main

import (
	"fmt"
	"net"

	"github.com/kpango/glg"
)

var UdpConnCache = make(map[string]net.Conn)

func ClientUdpInput(clientConn net.Conn, currentPacket *Packet) {
	defer HandlePanicError()
	clientAddr := fmt.Sprintf("%v:%v", currentPacket.ip4Header.sourceAddress.String(), currentPacket.udpHeader.sourcePort)
	remoteAddr := fmt.Sprintf("%v:%v", currentPacket.ip4Header.destinationAddress.String(), currentPacket.udpHeader.destinationPort)
	udpKey := fmt.Sprintf("%v->%v", clientAddr, remoteAddr)
	var outputChannel = UdpConnCache[udpKey]
	if outputChannel == nil {
		remoteConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   currentPacket.ip4Header.destinationAddress,
			Port: int(currentPacket.udpHeader.destinationPort),
		})
		if err != nil {
			glg.Errorf("open udp connection err=%v", err)
			return
		}
		UdpConnCache[udpKey] = remoteConn
		currentPacket.swapSourceAndDestination()
		outputChannel = remoteConn
		go RemoteUdpToClient(clientConn, remoteConn, currentPacket)
	}
	var headerLen = currentPacket.ip4Header.headerLength + UDP_HEADER_SIZE
	sendDataToRemote(outputChannel, headerLen, currentPacket.backingBuffer)
}

func RemoteUdpToClient(clientConn net.Conn, remoteConn *net.UDPConn, packet *Packet) {
	defer HandlePanicError()
	for {
		var buf = make([]byte, BUFFER_SIZE)
		n, _, err := remoteConn.ReadFromUDP(buf)
		if n > 0 {
			var headerLen = packet.ip4Header.headerLength + UDP_HEADER_SIZE
			var bytesBuf = NewBytesBuffer(BUFFER_SIZE)
			bytesBuf.putBytes(headerLen, buf[:n])
			packet.updateUDPBuffer(bytesBuf, int32(n))
			sendDataToClient(clientConn, bytesBuf)
		}
		if err != nil {
			glg.Errorf("open udp connection err=%v", err)
			break
		}
	}
}
