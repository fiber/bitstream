package bitstream

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/dgryski/go-bits"
	gryski "github.com/dgryski/go-bitstream"
)

type nullWriter struct{}

var bn = make([]byte, 32*1024, 32*1024)

func (nullWriter) Write(b []byte) (int, error) {
	if len(bn) < len(b) {
		bn = make([]byte, len(b)*2, len(b)*2)
	}
	copy(bn, b) // simulate a writer that actually does stuff
	return len(b), nil
}
func (nullWriter) WriteByte(b byte) error {
	return nil
}

const testWriteBits = 100 * 1024 * 1024

func BenchmarkWriteByteWriter001(b *testing.B) {
	w := NewByteWriter(nullWriter{})
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/2; j++ {
			w.WriteOne()
			w.WriteZero()
		}
		w.Flush()
	}
}

func BenchmarkWriteWriter001(b *testing.B) {
	w := NewWriter(nullWriter{})
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/2; j++ {
			w.WriteOne()
			w.WriteZero()
		}
		w.Flush()
	}
}

func BenchmarkGryskiWriter001(b *testing.B) {
	w := gryski.NewWriter(nullWriter{})
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/2; j++ {
			w.WriteBit(gryski.One)
			w.WriteBit(gryski.Zero)
		}
		w.Flush(gryski.Zero)
	}
}

func BenchmarkWriteWriter002(b *testing.B) {
	w := NewWriter(nullWriter{})
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/2; j++ {
			w.WriteBit(O)
			w.WriteBit(Z)
		}
		w.Flush()
	}
}

func BenchmarkWriteWriter003(b *testing.B) {
	w := NewWriter(nullWriter{})
	v := uint64(0x75)
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/5; j++ {
			w.WriteBits(v, 5)
		}
		w.Flush()
	}
}

func BenchmarkGryskiWriter003(b *testing.B) {
	w := gryski.NewWriter(nullWriter{})
	v := uint64(0x75)
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/5; j++ {
			w.WriteBits(v, 5)
		}
		w.Flush(gryski.Zero)
	}
}

func BenchmarkByteWriter004(b *testing.B) {
	w := NewByteWriter(nullWriter{})
	v := uint64(0x75757575)
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/25; j++ {
			w.WriteBits(v, 35)
		}
		w.Flush()
	}
}

func BenchmarkWriteWriter004(b *testing.B) {
	w := NewWriter(nullWriter{})
	v := uint64(0x75757575)
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/25; j++ {
			w.WriteBits(v, 35)
		}
		w.Flush()
	}
}

func BenchmarkGryskiWriter004(b *testing.B) {
	w := gryski.NewWriter(nullWriter{})
	v := uint64(0x75757575)
	for i := 0; i < b.N; i++ {
		for j := 0; j < testWriteBits/25; j++ {
			w.WriteBits(v, 35)
		}
		w.Flush(gryski.Zero)
	}
}

func BenchmarkReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rnd := rand.NewSource(1234)
		rd := NewReader(bytes.NewReader(random))
		for i := 0; i < 1000000; i++ {
			l, err := rd.ReadBits(6)
			if err != nil {
				b.Errorf("ReadBits returned error %v", err)
				return
			}
			n, err := rd.ReadBits(uint(l))
			if err != nil {
				b.Errorf("ReadBits returned error %v", err)
				return
			}
			rn := uint64(rnd.Int63())
			if n != rn {
				b.Errorf("random data decoding failed at pos %v, len %v, got %x, expected %x (diff %x)", i, l, n, rn, n-rn)
				return
			}
		}
	}
}

func BenchmarkGryskiReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rnd := rand.NewSource(1234)
		rd := gryski.NewReader(bytes.NewReader(random))
		for i := 0; i < 1000000; i++ {
			l, err := rd.ReadBits(6)
			if err != nil {
				b.Errorf("ReadBits returned error %v", err)
				return
			}
			n, err := rd.ReadBits(int(l))
			if err != nil {
				b.Errorf("ReadBits returned error %v", err)
				return
			}
			rn := uint64(rnd.Int63())
			if n != rn {
				b.Errorf("random data decoding failed at pos %v, len %v, got %x, expected %x (diff %x)", i, l, n, rn, n-rn)
				return
			}
		}
	}
}

var random = randomBytes()

func randomBytes() []byte {
	rnd := rand.NewSource(1234)
	w := new(bytes.Buffer)
	wr := NewWriter(w)
	for i := 0; i < 1000000; i++ {
		rn := uint64(rnd.Int63())
		l := 64 - bits.Clz(rn)
		if err := wr.WriteBits(l, 6); err != nil {
			panic(fmt.Errorf("WriteBits returned error %v", err))
		}
		if err := wr.WriteBits(rn, uint(l)); err != nil {
			panic(fmt.Errorf("WriteBits returned error %v", err))
		}
	}
	wr.Flush()
	return w.Bytes()
}
