package main

import (
	"net"
	"sync"
)

var (
	TCB_STATUS_SYN_SENT     int32 = 1
	TCB_STATUS_SYN_RECEIVED int32 = 2
	TCB_STATUS_ESTABLISHED  int32 = 3
	TCB_STATUS_CLOSE_WAIT   int32 = 4
	TCB_STATUS_LAST_ACK     int32 = 5
)

type TCB struct {
	TcbKey     string
	ClientAddr string
	RemoteAddr string

	//客户端的顺序码,每次发送多少数据就加多少,普通的无负载的数据包算做是1byte
	mySequenceNum, theirSequenceNum int64
	//客户端的ack码,为对方发来的seq码加上其发送的数据大小
	myAcknowledgementNum, theirAcknowledgementNum int64

	//记录发送的数据咯
	sendBytes int32
	tcbStatus int32

	//用来封装生成数据包
	referencePacket     *Packet
	ClientConn          net.Conn
	RemoteConn          net.Conn
	IsStartedRemoteRead bool
	Locker              sync.Locker
}

func NewTCB(tcbKey string, clientAddr string, remoteAddr string, clientConn net.Conn, referencePacket *Packet, mySequenceNum, theirSequenceNum, myAcknowledgementNum, theirAcknowledgementNum int64) *TCB {
	return &TCB{
		TcbKey:                  tcbKey,
		ClientAddr:              clientAddr,
		ClientConn:              clientConn,
		RemoteAddr:              remoteAddr,
		mySequenceNum:           mySequenceNum,
		theirSequenceNum:        theirSequenceNum,
		myAcknowledgementNum:    myAcknowledgementNum,
		theirAcknowledgementNum: theirAcknowledgementNum,
		sendBytes:               0,
		tcbStatus:               0,
		referencePacket:         referencePacket,
		Locker:                  &sync.Mutex{},
	}
}

func (this *TCB) calculateTransBytes(payloadSize int32) {
	this.sendBytes += payloadSize
}
