package main

import (
	"encoding/binary"
	"net"

	"github.com/kpango/glg"
	"golang.org/x/net/ipv4"
)

/**
 * Representation of an IP Packet
 * IP报头
 */
// TODO: Reduce public mutability

const BUFFER_SIZE int32 = 65535 // XXX: Is this ideal?

var (
	IP4_HEADER_SIZE int32 = 20
	TCP_HEADER_SIZE int32 = 20
	UDP_HEADER_SIZE int32 = 8
)

var (
	TcpHeaderFIN byte = 0x01
	TcpHeaderSYN byte = 0x02
	TcpHeaderRST byte = 0x04
	TcpHeaderPSH byte = 0x08
	TcpHeaderACK byte = 0x10
	TcpHeaderURG byte = 0x20
)

type TransportProtocol int32

const (
	TCP   TransportProtocol = 6
	UDP   TransportProtocol = 17
	Other TransportProtocol = 0xFF
)

// IP4报文头
type IP4Header struct {
	version                                 byte
	IHL                                     byte
	headerLength                            int32
	typeOfService                           int16
	totalLength                             int32
	identificationAndFlagsAndFragmentOffset int32
	TTL                                     int16
	protocolNum                             int16
	protocol                                TransportProtocol
	headerChecksum                          int32
	sourceAddress                           net.IP
	destinationAddress                      net.IP
	optionsAndPadding                       int32
}

func NewIP4Header(buffer *BytesBuffer) *IP4Header {
	ipHeader, err := ipv4.ParseHeader(buffer.getFullBytes())
	if err != nil {
		glg.Errorf("get ipv4 header err=%v", err)
		return nil
	}
	var this = &IP4Header{}
	var versionAndIHL = buffer.get(0)
	this.version = versionAndIHL >> 4
	this.IHL = versionAndIHL & 0x0F
	this.headerLength = int32(this.IHL << 2)
	this.typeOfService = getUnsignedByte(buffer.get(1))
	this.totalLength = getUnsignedShort(buffer.getShort(2))

	this.identificationAndFlagsAndFragmentOffset = buffer.getInt(4)

	this.TTL = getUnsignedByte(buffer.get(8))
	this.protocolNum = getUnsignedByte(buffer.get(9))
	this.protocol = TransportProtocol(this.protocolNum)
	this.headerChecksum = getUnsignedShort(buffer.getShort(10))

	var sourceAddressBytes = make([]byte, 4)
	buffer.getBytes(12, sourceAddressBytes)
	this.sourceAddress = sourceAddressBytes

	var destinationAddressBytes = make([]byte, 4)
	buffer.getBytes(16, destinationAddressBytes)
	this.destinationAddress = destinationAddressBytes
	glg.Logf("get ipv4 header length=%v,ipHeader.Len=%v,myProtocal=%v,thire protocal=%v", this.headerLength, ipHeader.Len, int32(this.protocol), ipHeader.Protocol)
	return this
}

func (this *IP4Header) fillHeader(buffer *BytesBuffer) {
	buffer.put(0, this.version<<4|this.IHL)
	buffer.put(1, byte(this.typeOfService))
	buffer.putShort(2, int16(this.totalLength))

	buffer.putInt(4, this.identificationAndFlagsAndFragmentOffset)

	buffer.put(8, byte(this.TTL))
	buffer.put(9, byte(this.protocol))
	buffer.putShort(10, int16(this.headerChecksum))

	buffer.putBytes(12, this.sourceAddress)
	buffer.putBytes(16, this.destinationAddress)
}

type TCPHeader struct {
	sourcePort            int32
	destinationPort       int32
	sequenceNumber        int64
	acknowledgementNumber int64
	dataOffsetAndReserved byte
	headerLength          int32
	flags                 byte
	window                int32
	checksum              int32
	urgentPointer         int32
	optionsAndPadding     []byte
}

func NewTCPHeader(buffer *BytesBuffer, ip4HeaderLen int32) *TCPHeader {
	var this = &TCPHeader{}
	this.sourcePort = getUnsignedShort(buffer.getShort(ip4HeaderLen + 0))
	this.destinationPort = getUnsignedShort(buffer.getShort(ip4HeaderLen + 2))

	this.sequenceNumber = getUnsignedInt(buffer.getInt(ip4HeaderLen + 4))
	this.acknowledgementNumber = getUnsignedInt(buffer.getInt(ip4HeaderLen + 8))

	this.dataOffsetAndReserved = buffer.get(ip4HeaderLen + 12)
	this.headerLength = int32((this.dataOffsetAndReserved & 0xF0) >> 2)
	this.flags = buffer.get(ip4HeaderLen + 13)
	this.window = getUnsignedShort(buffer.getShort(ip4HeaderLen + 14))

	this.checksum = getUnsignedShort(buffer.getShort(ip4HeaderLen + 16))
	this.urgentPointer = getUnsignedShort(buffer.getShort(ip4HeaderLen + 18))

	var optionsLength = this.headerLength - TCP_HEADER_SIZE
	if optionsLength > 0 {
		this.optionsAndPadding = make([]byte, optionsLength)
		buffer.getBytes(ip4HeaderLen+20, this.optionsAndPadding)
	}
	return this
}

func (this *TCPHeader) isFIN() bool {
	return this.flags&TcpHeaderFIN == TcpHeaderFIN
}

func (this *TCPHeader) isSYN() bool {
	return this.flags&TcpHeaderSYN == TcpHeaderSYN
}

func (this *TCPHeader) isRST() bool {
	return this.flags&TcpHeaderRST == TcpHeaderRST
}

func (this *TCPHeader) isPSH() bool {
	return this.flags&TcpHeaderPSH == TcpHeaderPSH
}

func (this *TCPHeader) isACK() bool {
	return this.flags&TcpHeaderACK == TcpHeaderACK
}

func (this *TCPHeader) isURG() bool {
	return this.flags&TcpHeaderURG == TcpHeaderURG
}

func (this *TCPHeader) fillHeader(buffer *BytesBuffer, ip4HeaderLen int32) {
	buffer.putShort(ip4HeaderLen+0, int16(this.sourcePort))
	buffer.putShort(ip4HeaderLen+2, int16(this.destinationPort))

	buffer.putInt(ip4HeaderLen+4, int32(this.sequenceNumber))
	buffer.putInt(ip4HeaderLen+8, int32(this.acknowledgementNumber))

	buffer.put(ip4HeaderLen+12, this.dataOffsetAndReserved)
	buffer.put(ip4HeaderLen+13, this.flags)
	buffer.putShort(ip4HeaderLen+14, int16(this.window))

	buffer.putShort(ip4HeaderLen+16, int16(this.checksum))
	buffer.putShort(ip4HeaderLen+18, int16(this.urgentPointer))
}

type UDPHeader struct {
	sourcePort      int32
	destinationPort int32
	length          int32
	checksum        int32
}

func NewUDPHeader(buffer *BytesBuffer, ip4HeaderLen int32) *UDPHeader {
	var this = &UDPHeader{}
	this.sourcePort = getUnsignedShort(buffer.getShort(ip4HeaderLen + 0))
	this.destinationPort = getUnsignedShort(buffer.getShort(ip4HeaderLen + 2))

	this.length = getUnsignedShort(buffer.getShort(ip4HeaderLen + 4))
	this.checksum = getUnsignedShort(buffer.getShort(ip4HeaderLen + 6))
	return this
}

func (this *UDPHeader) fillHeader(buffer *BytesBuffer, ip4HeaderLen int32) {
	buffer.putShort(ip4HeaderLen+0, int16(this.sourcePort))
	buffer.putShort(ip4HeaderLen+2, int16(this.destinationPort))

	buffer.putShort(ip4HeaderLen+4, int16(this.length))
	buffer.putShort(ip4HeaderLen+6, int16(this.checksum))
}

type Packet struct {
	//IP TCP报文头
	ip4Header *IP4Header
	tcpHeader *TCPHeader
	udpHeader *UDPHeader
	//实际的传输数据
	backingBuffer *BytesBuffer
	isTCP         bool
	isUDP         bool
}

func NewPacket(buffer *BytesBuffer) *Packet {
	var this = &Packet{}
	this.ip4Header = NewIP4Header(buffer)
	if this.ip4Header.protocol == TCP {
		this.tcpHeader = NewTCPHeader(buffer, this.ip4Header.headerLength)
		this.isTCP = true
	} else if this.ip4Header.protocol == UDP {
		this.udpHeader = NewUDPHeader(buffer, this.ip4Header.headerLength)
		this.isUDP = true
	}
	this.backingBuffer = buffer
	return this
}

/**
 * 获取payloadSize的大小
 * @return
 */
func (p *Packet) getPureDataSize() int32 {
	if p.backingBuffer != nil {
		headerLen := p.ip4Header.headerLength + TCP_HEADER_SIZE
		return p.backingBuffer.Limit - headerLen
	}
	return 0
}

// 调换源和目的地址
func (this *Packet) swapSourceAndDestination() {
	var newSourceAddress = this.ip4Header.destinationAddress
	this.ip4Header.destinationAddress = this.ip4Header.sourceAddress
	this.ip4Header.sourceAddress = newSourceAddress
	if this.isUDP {
		var newSourcePort = this.udpHeader.destinationPort
		this.udpHeader.destinationPort = this.udpHeader.sourcePort
		this.udpHeader.sourcePort = newSourcePort
	} else if this.isTCP {
		var newSourcePort = this.tcpHeader.destinationPort
		this.tcpHeader.destinationPort = this.tcpHeader.sourcePort
		this.tcpHeader.sourcePort = newSourcePort
	}
}

// position从0开始写入到40 buffer 40后的为实际的数据
func (this *Packet) updateTCPBuffer(buffer *BytesBuffer, flags byte, sequenceNum int64, ackNum int64, pureDataSize int32) {
	this.fillHeader(buffer)
	ip4HeaderLen := this.ip4Header.headerLength
	this.backingBuffer = buffer

	this.tcpHeader.flags = flags
	this.backingBuffer.put(ip4HeaderLen+13, flags)

	this.tcpHeader.sequenceNumber = sequenceNum
	this.backingBuffer.putInt(ip4HeaderLen+4, int32(sequenceNum))

	this.tcpHeader.acknowledgementNumber = ackNum
	this.backingBuffer.putInt(ip4HeaderLen+8, int32(ackNum))

	// Reset header size, since we don't need options
	var dataOffset = byte(ip4HeaderLen << 2)
	this.tcpHeader.dataOffsetAndReserved = dataOffset
	this.backingBuffer.put(ip4HeaderLen+12, dataOffset)
	this.updateTCPChecksum(pureDataSize)
	var ip4TotalLength = ip4HeaderLen + TCP_HEADER_SIZE + pureDataSize
	this.backingBuffer.putShort(2, int16(ip4TotalLength))
	this.ip4Header.totalLength = ip4TotalLength

	this.updateIP4Checksum()
}

func (this *Packet) updateUDPBuffer(buffer *BytesBuffer, payloadSize int32) {
	this.fillHeader(buffer)
	ip4HeaderLen := this.ip4Header.headerLength
	this.backingBuffer = buffer
	var udpTotalLength = UDP_HEADER_SIZE + payloadSize
	this.backingBuffer.putShort(ip4HeaderLen+4, int16(udpTotalLength))
	this.udpHeader.length = udpTotalLength
	// Disable UDP checksum validation
	this.backingBuffer.putShort(ip4HeaderLen+6, 0)
	this.udpHeader.checksum = 0
	var ip4TotalLength = ip4HeaderLen + udpTotalLength
	this.backingBuffer.putShort(2, int16(ip4TotalLength))
	this.ip4Header.totalLength = ip4TotalLength
	this.updateIP4Checksum()
}

func (this *Packet) updateIP4Checksum() {
	// Clear previous checksum
	this.backingBuffer.putShort(10, 0)

	var ipLength = this.ip4Header.headerLength
	var sum int32 = 0
	var readLen int32 = 0
	for ipLength > 0 {
		sum += getUnsignedShort(this.backingBuffer.getShort(readLen))
		ipLength -= 2
		readLen += 2
	}
	for sum>>16 > 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	sum = ^sum
	this.ip4Header.headerChecksum = sum
	this.backingBuffer.putShort(10, int16(sum))
}

func (this *Packet) updateTCPChecksum(pureDataSize int32) {
	var sum int32 = 0
	ip4HeaderLen := this.ip4Header.headerLength
	var tcpLength = TCP_HEADER_SIZE + pureDataSize

	// Calculate pseudo-header checksum
	var buffer = BytesBufferWrap(this.ip4Header.sourceAddress)
	sum = getUnsignedShort(buffer.getShort(0)) + getUnsignedShort(buffer.getShort(2))

	buffer = BytesBufferWrap(this.ip4Header.destinationAddress)
	sum += getUnsignedShort(buffer.getShort(0)) + getUnsignedShort(buffer.getShort(2))

	sum += int32(TCP) + tcpLength

	// Clear previous checksum
	this.backingBuffer.putShort(ip4HeaderLen+16, 0)

	// Calculate TCP segment checksum
	var readLen int32 = 0
	for tcpLength > 1 {
		sum += getUnsignedShort(this.backingBuffer.getShort(ip4HeaderLen + readLen))
		tcpLength -= 2
		readLen += 2
	}
	if tcpLength > 0 {
		sum += int32(getUnsignedByte(this.backingBuffer.get(ip4HeaderLen+readLen)) << 8)
	}

	for sum>>16 > 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	sum = ^sum
	this.tcpHeader.checksum = sum
	this.backingBuffer.putShort(ip4HeaderLen+16, int16(sum))
}

func (this *Packet) fillHeader(buffer *BytesBuffer) {
	this.ip4Header.fillHeader(buffer)
	if this.isUDP {
		this.udpHeader.fillHeader(buffer, this.ip4Header.headerLength)
	} else if this.isTCP {
		this.tcpHeader.fillHeader(buffer, this.ip4Header.headerLength)
	}
}

func getUnsignedByte(value byte) int16 {
	return int16(value) & 0xFF
}

func getUnsignedShort(value int16) int32 {
	return int32(value) & 0xFFFF
}

func getUnsignedInt(value int32) int64 {
	return int64(value) & 0xFFFFFFFF
}

func Int16ToBytes(v uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)
	return buf
}

func Int32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	return buf
}

func Int64ToBytes(v uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)
	return buf
}

func BytesToInt16(v []byte) uint16 {
	return binary.BigEndian.Uint16(v)
}

func BytesToInt32(v []byte) uint32 {
	return binary.BigEndian.Uint32(v)
}

func BytesToInt64(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}
