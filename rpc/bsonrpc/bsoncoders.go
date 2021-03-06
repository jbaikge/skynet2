package bsonrpc

import (
	"bufio"
	"fmt"
	"io"
	"labix.org/v2/mgo/bson"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(v interface{}) (err error) {
	buf, err := bson.Marshal(v)
	if err != nil {
		return
	}
	n, err := e.w.Write(buf)
	if err != nil {
		return
	}
	if l := len(buf); n != l {
		return fmt.Errorf("Wrote %d bytes, should have wrote %d", n, l)
	}
	return
}

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	buf := bufio.NewReader(r)
	return &Decoder{r: buf}
}

func (d *Decoder) Decode(pv interface{}) (err error) {
	var lbuf [4]byte
	n, err := d.r.Read(lbuf[:])
	if n == 0 {
		return io.EOF
	}
	if n != 4 {
		return fmt.Errorf("Corrupted BSON stream: could only read %d", n)
	}
	if err != nil {
		return
	}

	length := (int(lbuf[0]) << 0) |
		(int(lbuf[1]) << 8) |
		(int(lbuf[2]) << 16) |
		(int(lbuf[3]) << 24)

	buf := make([]byte, length)
	copy(buf[0:4], lbuf[:])
	n, err = io.ReadFull(d.r, buf[4:])
	if err != nil {
		return
	}
	if n+4 != length {
		return fmt.Errorf("Expected %d bytes, read %d", length, n)
	}

	err = bson.Unmarshal(buf, pv)

	return
}
