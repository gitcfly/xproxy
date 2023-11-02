package main

type BytesBuffer struct {
	Data     []byte
	Limit    int32
	Capacity int32
}

func NewBytesBuffer(size int32) *BytesBuffer {
	var data = make([]byte, size)
	return &BytesBuffer{
		Data:     data,
		Limit:    0,
		Capacity: size,
	}
}

func (r *BytesBuffer) limit() int32 {
	return r.Limit
}

func (r *BytesBuffer) clear() {
	r.Limit = 0
	r.Capacity = 0
	r.Data = []byte{}
}

func (r *BytesBuffer) get(at int32) byte {
	return r.Data[at]
}

func (r *BytesBuffer) getShort(at int32) int16 {
	return BytesToInt16(r.Data[at : at+2])
}

func (r *BytesBuffer) getInt(at int32) int32 {
	return BytesToInt32(r.Data[at : at+4])
}

func (r *BytesBuffer) getInt64(at int32) int64 {
	return BytesToInt64(r.Data[at : at+8])
}

func (r *BytesBuffer) getBytes(at int32, dest []byte) int32 {
	readLen := r.Limit - at
	if readLen > int32(len(dest)) {
		readLen = int32(len(dest))
	}
	for i := 0; int32(i) < readLen; i++ {
		dest[i] = r.Data[at+int32(i)]
	}
	return readLen
}

func (r *BytesBuffer) getFullBytes() []byte {
	return r.Data[:r.Limit]
}

func (r *BytesBuffer) put(at int32, v byte) {
	r.Data[at] = v
	r.Limit = MaxInt32(r.Limit, at+1)
}

func (r *BytesBuffer) putShort(at int32, v int16) {
	vs := Int16ToBytes(v)
	for i := 0; i < len(vs); i++ {
		r.Data[at+int32(i)] = vs[i]
	}
	r.Limit = MaxInt32(r.Limit, at+2)
}

func (r *BytesBuffer) putInt(at int32, v int32) {
	vs := Int32ToBytes(v)
	for i := 0; i < len(vs); i++ {
		r.Data[at+int32(i)] = vs[i]
	}
	r.Limit = MaxInt32(r.Limit, at+4)
}

func (r *BytesBuffer) putInt64(at int32, v int64) {
	vs := Int64ToBytes(v)
	for i := 0; i < len(vs); i++ {
		r.Data[at+int32(i)] = vs[i]
	}
	r.Limit = MaxInt32(r.Limit, at+8)
}

func (r *BytesBuffer) putBytes(at int32, dest []byte) {
	for i := 0; i < len(dest); i++ {
		r.Data[at+int32(i)] = dest[i]
	}
	r.Limit = MaxInt32(r.Limit, at+int32(len(dest)))
}

func BytesBufferWrap(vs []byte) *BytesBuffer {
	return &BytesBuffer{
		Data:     vs,
		Limit:    int32(len(vs)),
		Capacity: int32(len(vs)),
	}
}

func MaxInt32(v1 int32, v2 int32) int32 {
	if v1 > v2 {
		return v1
	}
	return v2
}
