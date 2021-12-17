package oscar

import (
	"encoding"
	"encoding/binary"
)

type Buffer struct {
	d []byte
}

func (b *Buffer) WriteUint8(x uint8) {
	b.d = append(b.d, x)
}

func (b *Buffer) WriteUint16(x uint16) {
	b.d = append(b.d, 0, 0)
	binary.BigEndian.PutUint16(b.d[len(b.d)-2:], x)
}

func (b *Buffer) WriteUint32(x uint32) {
	b.d = append(b.d, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(b.d[len(b.d)-4:], x)
}

func (b *Buffer) WriteUint64(x uint64) {
	b.d = append(b.d, 0, 0, 0, 0, 0, 0, 0, 0)
	binary.BigEndian.PutUint64(b.d[len(b.d)-8:], x)
}

func (b *Buffer) WriteString(x string) {
	b.d = append(b.d, []byte(x)...)
}

func (b *Buffer) Write(x []byte) (int, error) {
	b.d = append(b.d, x...)
	return len(x), nil
}

func (b *Buffer) WriteBinary(e encoding.BinaryMarshaler) {
	d, err := e.MarshalBinary()
	panicIfError(err)
	b.d = append(b.d, d...)
}

func (b *Buffer) Bytes() []byte {
	return b.d
}
