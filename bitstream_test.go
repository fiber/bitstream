package bitstream

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"

	bits "github.com/dgryski/go-bits"
)

func TestBitStream001(t *testing.T) {
	br := NewReader(strings.NewReader("1"))
	b, err := br.ReadByte()
	if err != nil {
		t.Errorf("ReadByte returned error %v", err)
	}
	if b != '1' {
		t.Error("ReadByte returned invalid content")
	}
	b, err = br.ReadByte()
	if err != io.EOF {
		t.Error("ReadByte did not return expected EOF")
	}
}

func TestBitStream002(t *testing.T) {
	br := NewReader(strings.NewReader("bar"))
	by := make([]byte, 3)
	err := br.ReadBytes(by)
	if err != nil {
		t.Errorf("ReadByte returned error %v", err)
	}
	if by[0] != 'b' || by[1] != 'a' || by[2] != 'r' {
		t.Error("ReadByte returned invalid content")
	}
	if _, err = br.ReadByte(); err != io.EOF {
		t.Error("ReadByte did not return expected EOF")
	}
}

func TestBitStream003(t *testing.T) {
	w := new(bytes.Buffer)
	wr := NewWriter(w)
	b := []byte("Hello World!")
	wr.Write(b)
	wr.Flush()
	wb := w.Bytes()
	r := bytes.NewReader(wb)
	if string(wb) != "Hello World!" {
		t.Errorf("Writer encoded invalid data")
	}
	rd := NewReader(r)
	if err := rd.ReadBytes(b); err != nil {
		t.Errorf("ReadBit returned error %v", err)
	}
	if string(b) != "Hello World!" {
		t.Error("ReadBit returned invalid content")
	}
	if _, err := rd.ReadBit(); err != io.EOF {
		t.Error("ReadBit did not return expected EOF")
	}
}

func TestBitStream004(t *testing.T) {
	w := new(bytes.Buffer)
	wr := NewWriter(w)
	wr.WriteOne()
	wr.WriteZero()
	b := []byte("foo bar baz!")
	wr.Write(b)
	wr.Flush()
	r := bytes.NewReader(w.Bytes())
	rd := NewReader(r)
	bit, err := rd.ReadBit()
	if err != nil {
		t.Errorf("ReadBit returned error %v", err)
	}
	if !bit {
		t.Error("ReadBit returned invalid bit")
	}
	bit, err = rd.ReadBit()
	if err != nil {
		t.Errorf("ReadBit returned error %v", err)
	}
	if bit {
		t.Error("ReadBit returned invalid bit")
	}
	if err = rd.ReadBytes(b); err != nil {
		t.Errorf("ReadBit returned error %v", err)
	}
	if string(b) != "foo bar baz!" {
		t.Error("ReadBit returned invalid content")
	}
	if _, err := rd.ReadByte(); err != io.EOF {
		t.Error("ReadByte did not return expected EOF")
	}
}

func TestBitStream006(t *testing.T) {
	rnd := rand.NewSource(1234)
	w := new(bytes.Buffer)
	wr := NewWriter(w)
	for i := 0; i < 1000000; i++ {
		rn := uint64(rnd.Int63())
		l := 64 - bits.Clz(rn)
		if err := wr.WriteBits(l, 6); err != nil {
			t.Errorf("WriteBits returned error %v", err)
			return
		}
		if err := wr.WriteBits(rn, uint(l)); err != nil {
			t.Errorf("WriteBits returned error %v", err)
			return
		}
	}
	wr.Flush()
	wb := w.Bytes()
	r := bytes.NewReader(wb)
	fmt.Printf("random data encoded length is %v\n", len(wb))
	rd := NewReader(r)
	rnd = rand.NewSource(1234)
	for i := 0; i < 1000000; i++ {
		l, err := rd.ReadBits(6)
		if err != nil {
			t.Errorf("ReadBits returned error %v", err)
			return
		}
		n, err := rd.ReadBits(uint(l))
		if err != nil {
			t.Errorf("ReadBits returned error %v", err)
			return
		}
		rn := uint64(rnd.Int63())
		if n != rn {
			t.Errorf("random data decoding failed at pos %v, len %v, got %x, expected %x (diff %x)", i, l, n, rn, n-rn)
			return
		}
	}
	if _, err := rd.ReadByte(); err != io.EOF {
		t.Error("ReadByte did not return expected EOF")
		return
	}
}
