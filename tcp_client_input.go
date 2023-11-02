package main

import (
	"fmt"
	"math"
	"math/rand"
	"net"

	"github.com/kpango/glg"
)

// 本地设备发送到远端网络的数据处理

func ClientTcpInput(clientConn net.Conn, packet *Packet) {
	defer HandlePanicError()
	var clientDataBuffer = packet.backingBuffer
	//clientDataBuffer 其实也是可复用的吧
	packet.backingBuffer = nil
	var newRspBuffer = NewBytesBuffer(BUFFER_SIZE)

	clientAddr := fmt.Sprintf("%v:%v", packet.ip4Header.sourceAddress.String(), packet.tcpHeader.sourcePort)
	remoteAddr := fmt.Sprintf("%v:%v", packet.ip4Header.destinationAddress.String(), packet.tcpHeader.destinationPort)
	tcbKey := fmt.Sprintf("%v->%v", clientAddr, remoteAddr)
	glg.Infof("ClientTcpInput tcbKey=%v", tcbKey)
	tcb := tcbMap[tcbKey]
	if tcb == nil {
		InitTcp(tcbKey, clientAddr, remoteAddr, clientConn, packet, newRspBuffer)
		return
	}
	if packet.tcpHeader.isSYN() { // 握手
		dealDuplicatedSYN(tcb, tcbKey, packet, newRspBuffer)
		return
	}
	if packet.tcpHeader.isFIN() {
		finishConnect(tcb, tcbKey, packet, newRspBuffer)
		return
	}
	if packet.tcpHeader.isACK() {
		transData(tcbKey, tcb, packet, clientDataBuffer, newRspBuffer)
	}
	clientDataBuffer.clear()
}

func InitTcp(tcbKey string, clientAddr, remoteAddr string, clientConn net.Conn, referencePacket *Packet, newRspBuffer *BytesBuffer) {
	//TODO 抓包咯,这样很不妥呀,还要根据相关信息构建包orz

	var newEmptyBuffer = NewBytesBuffer(BUFFER_SIZE)
	referencePacket.updateTCPBuffer(newEmptyBuffer, TcpHeaderSYN, referencePacket.tcpHeader.sequenceNumber,
		referencePacket.tcpHeader.acknowledgementNumber, 0)
	newEmptyBuffer.clear()

	referencePacket.swapSourceAndDestination()
	var tcb = NewTCB(tcbKey, clientAddr, remoteAddr, clientConn, referencePacket, int64(rand.Int31n(math.MaxInt16)), referencePacket.tcpHeader.sequenceNumber,
		referencePacket.tcpHeader.sequenceNumber+1, referencePacket.tcpHeader.acknowledgementNumber)
	tcbMap[tcbKey] = tcb
	remoteConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		glg.Errorf("net.Dial remote err=%v", err)
		tcb.tcbStatus = TCB_STATUS_SYN_SENT
		newRspBuffer.clear()
		return
	}
	tcb.RemoteConn = remoteConn
	tcb.tcbStatus = TCB_STATUS_SYN_RECEIVED
	referencePacket.updateTCPBuffer(newRspBuffer, TcpHeaderSYN|TcpHeaderACK, tcb.mySequenceNum, tcb.myAcknowledgementNum, 0)
	tcb.mySequenceNum++
	sendDataToClient(clientConn, newRspBuffer) // 需要发送数据到客户端
	go RemoteTcpInput(tcb)
}

func dealDuplicatedSYN(tcb *TCB, tcbKey string, currentPacket *Packet, newRspBuffer *BytesBuffer) {
	tcb.Locker.Lock()
	defer tcb.Locker.Unlock()
	if tcb.tcbStatus == TCB_STATUS_SYN_SENT {
		//如果是SYN发送的状态,即还在与远程服务器建立连接ing..
		tcb.myAcknowledgementNum = currentPacket.tcpHeader.sequenceNumber + 1
		return
	}
	sendRST(tcb, tcbKey, 1, newRspBuffer)
}

// 连接重置咯,即断开之前的连接
func sendRST(tcb *TCB, tcbKey string, prevPayloadSize int32, newRspBuffer *BytesBuffer) {
	tcb.referencePacket.updateTCPBuffer(newRspBuffer, TcpHeaderRST, 0, tcb.myAcknowledgementNum+int64(prevPayloadSize), 0)
	sendDataToClient(tcb.ClientConn, newRspBuffer)
	delete(tcbMap, tcbKey)
}

func finishConnect(tcb *TCB, tcbKey string, currentPacket *Packet, newRspBuffer *BytesBuffer) {
	tcb.Locker.Lock()
	defer tcb.Locker.Unlock()
	//标识已经被改变咯
	tcb.myAcknowledgementNum = currentPacket.tcpHeader.sequenceNumber + 1
	tcb.tcbStatus = TCB_STATUS_LAST_ACK
	currentPacket.updateTCPBuffer(newRspBuffer, TcpHeaderFIN|TcpHeaderACK, tcb.mySequenceNum, tcb.myAcknowledgementNum, 0)
	tcb.mySequenceNum++
	// TcpHeaderFIN counts as a byte
	delete(tcbMap, tcbKey)
	sendDataToClient(tcb.ClientConn, newRspBuffer)
}

/**
 * 传递实际的数据
 *
 * @param tcb
 */
func transData(tcbKey string, tcb *TCB, currentPacket *Packet, clientDataBuffer *BytesBuffer, newRspBuffer *BytesBuffer) {
	//1.发送ACK码 2.传递真实数据
	var ipAndTcpHeaderLength = currentPacket.ip4Header.headerLength + TCP_HEADER_SIZE
	var payloadSize = clientDataBuffer.limit() - ipAndTcpHeaderLength
	tcb.Locker.Lock()
	defer tcb.Locker.Unlock()
	if tcb.tcbStatus == TCB_STATUS_LAST_ACK {
		delete(tcbMap, tcbKey)
		return
	}
	//无数据的直接ignore了
	if payloadSize == 0 {
		return
	}
	//发送完数据咯,那么就执行真正的数据访问
	if tcb.RemoteConn == nil {
		glg.Error("remote connection 为 null")
		return
	}
	//监听读的状态咯
	if tcb.tcbStatus == TCB_STATUS_SYN_RECEIVED {
		tcb.tcbStatus = TCB_STATUS_ESTABLISHED
	} else if tcb.tcbStatus == TCB_STATUS_ESTABLISHED {
		glg.Infof("establish ing")
	} else {
		glg.Infof("连接还没建立好")
		return
	}
	err := sendDataToRemote(tcb.RemoteConn, ipAndTcpHeaderLength, clientDataBuffer)
	if err != nil {
		glg.Errorf("传输错误了，err=%v", err)
		sendRST(tcb, tcbKey, payloadSize, newRspBuffer)
		return
	}
	//记录发送数据
	tcb.calculateTransBytes(payloadSize)
	currentPacket.swapSourceAndDestination()
	tcb.myAcknowledgementNum = currentPacket.tcpHeader.sequenceNumber + int64(payloadSize)
	currentPacket.updateTCPBuffer(newRspBuffer, TcpHeaderACK, tcb.mySequenceNum, tcb.myAcknowledgementNum, 0)
	sendDataToClient(tcb.ClientConn, newRspBuffer)
}
