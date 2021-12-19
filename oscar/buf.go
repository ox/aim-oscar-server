package oscar

import (
	"aim-oscar/util"
	"encoding"
	"encoding/binary"
	"io"
)

// Buffer is a handy byte slice that reads from the front and writes on the end
type Buffer struct {
	d []byte
}

// Seek moves the read cursor forward. If the cursor is beyond the length of the byte
// slice then the buffer slice is just replaced with an empty slice
func (b *Buffer) Seek(n int) {
	if n > len(b.d) {
		b.d = make([]byte, 0)
		return
	}
	b.d = b.d[n:]
}

func (b *Buffer) Read(d []byte) (int, error) {
	if len(d) > len(b.d) {
		return 0, io.EOF
	}

	n := copy(d, b.d[0:len(d)])
	return n, nil
}

func (b *Buffer) ReadUint8() (uint8, error) {
	if len(b.d) < 1 {
		return 0, io.EOF
	}
	ret := uint8(b.d[0])
	b.d = b.d[1:]
	return ret, nil
}

func (b *Buffer) ReadUint16() (uint16, error) {
	if len(b.d) < 2 {
		return 0, io.EOF
	}
	ret := binary.BigEndian.Uint16(b.d[0:2])
	b.d = b.d[2:]
	return ret, nil
}

func (b *Buffer) ReadUint32() (uint32, error) {
	if len(b.d) < 4 {
		return 0, io.EOF
	}
	ret := binary.BigEndian.Uint32(b.d[0:4])
	b.d = b.d[4:]
	return ret, nil
}

func (b *Buffer) ReadUint64() (uint64, error) {
	if len(b.d) < 8 {
		return 0, io.EOF
	}
	ret := binary.BigEndian.Uint64(b.d[0:8])
	b.d = b.d[8:]
	return ret, nil
}

// ReadLPString reads a length-prefixed string. The first byte should be the string length
// followed by that many bytes. Returns io.EOF if there are less bytes than indicated.
func (b *Buffer) ReadLPString() (string, error) {
	length, err := b.ReadUint8()
	if err != nil {
		return "", nil
	}

	if len(b.d) < int(length) {
		return "", io.EOF
	}

	str := string(b.d[:length])
	b.d = b.d[length:]
	return str, nil
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

func (b *Buffer) WriteLPString(x string) {
	b.WriteUint8(uint8(len(x)))
	b.WriteString(x)
}

func (b *Buffer) Write(x []byte) (int, error) {
	b.d = append(b.d, x...)
	return len(x), nil
}

func (b *Buffer) WriteBinary(e encoding.BinaryMarshaler) {
	d, err := e.MarshalBinary()
	util.PanicIfError(err)
	b.d = append(b.d, d...)
}

func (b *Buffer) Bytes() []byte {
	return b.d
}
