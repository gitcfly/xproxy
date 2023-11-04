package main

import (
	"github.com/kpango/glg"
)

// 网络发送到客户端的数据处理
func RemoteTcpInput(tcb *TCB) {
	defer HandlePanicError()
	for {
		buf := make([]byte, BUFFER_SIZE)
		readLen, err := tcb.RemoteConn.Read(buf)
		transDataToClient(buf, int32(readLen), tcb)
		if err != nil || readLen <= 0 {
			glg.Errorf("remote read data err=%v", err)
			return
		}
	}
}

func transDataToClient(data []byte, readBytes int32, tcb *TCB) {
	var newRspBuffer = NewBytesBuffer(BUFFER_SIZE)
	//这样实际数据就能写道里面去了
	tcb.Locker.Lock()
	defer tcb.Locker.Unlock()
	var responsePacket = tcb.referencePacket
	tcpIpHeaderLength := responsePacket.ip4Header.headerLength + TCP_HEADER_SIZE
	//Log.logd(this, "获取回来的数据大小为" + readBytes);
	if readBytes <= 0 {
		tcb.RemoteConn.Close()
		tcb.tcbStatus = TCB_STATUS_LAST_ACK
		responsePacket.updateTCPBuffer(newRspBuffer, TcpHeaderFIN|TcpHeaderACK, tcb.mySequenceNum, tcb.myAcknowledgementNum, 0)
		tcb.mySequenceNum++
		sendDataToClient(tcb.ClientConn, newRspBuffer)
		tcb.ClientConn.Close()
		//Log.logd(this, "数据读取完毕");
		return
	}
	newRspBuffer.putBytes(tcpIpHeaderLength, data[:readBytes])
	tcb.calculateTransBytes(readBytes)
	responsePacket.updateTCPBuffer(newRspBuffer, TcpHeaderACK|TcpHeaderPSH, tcb.mySequenceNum, tcb.myAcknowledgementNum, readBytes)
	tcb.mySequenceNum = tcb.mySequenceNum + int64(readBytes)
	//TODO 这个真让人疑惑
	sendDataToClient(tcb.ClientConn, newRspBuffer)
}
