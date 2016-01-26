package bitstream

import "io"

type (
	// Reader writes bits to an underlying reader
	Reader struct {
		r  io.Reader
		b  byte
		p  uint32
		br io.ByteReader
	}
)

// NewReader creates a new reader
func NewReader(r io.Reader) *Reader {
	rdr := &Reader{r: r}
	if br, ok := r.(io.ByteReader); ok {
		rdr.br = br
	}
	return rdr
}

// ReadBit reads a single bit
func (r *Reader) ReadBit() (Bit, error) {
	if r.p > 0 {
		d := (r.b & 0x80)
		r.p--
		r.b <<= 1
		return d != 0, nil
	}
	var b [1]byte
	if n, err := r.r.Read(b[:]); n != 1 || err != nil {
		return Zero, err
	}
	b0 := b[0]
	r.p = 7
	r.b = b0 << 1
	return (b0 & 0x80) != 0, nil
}

// ReadByte reads a single byte from the stream
func (r *Reader) ReadByte() (byte, error) {
	var b [1]byte
	n, err := r.r.Read(b[:])
	if n != 1 || err != nil {
		return 0, err
	}
	p := r.p
	if p == 0 {
		return b[0], nil
	}
	by := r.b | b[0]>>p
	r.b = b[0] << (8 - p)
	return by, nil
}

// ReadBytes reads len(b) bytes from the stream - it will return an error if not
// enough bytes are available
func (r *Reader) ReadBytes(b []byte) error {
	// we use b as our buffer to read enough bytes from from the reader
	n, err := r.r.Read(b)
	if err != nil {
		return err
	}
	if n < len(b) {
		if err != nil {
			return err
		}
		for n < len(b) {
			nn, err := r.r.Read(b[n:])
			if err != nil {
				return err
			}
			n += nn
		}
	}
	p := r.p
	if p == 0 {
		return nil
	}
	p8 := 8 - p
	by := r.b
	for i, c := range b {
		b[i] = by | b[i]>>p
		by = c << p8
	}
	r.b = by
	return nil
}

// ReadBits reads n bits from the stream
func (r *Reader) ReadBits(n uint) (uint64, error) {
	var (
		u uint64
	)
	l := n / 8
	if l > 0 {
		var b [8]byte
		if err := r.ReadBytes(b[:l]); err != nil {
			return 0, err
		}
		//u = uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 | uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
		//u >>= (8 * (8 - l))
		for p := uint(0); p < l; p++ {
			u = (u << 8) | uint64(b[p])
		}
		n -= l * 8
	}
	for ; n > 0; n-- {
		// TODO(se): we can change this to 1-2 mask and max one read
		bit, err := r.ReadBit()
		if err != nil {
			return 0, err
		}
		if bit {
			u = (u << 1) | 1
		} else {
			u <<= 1
		}
	}
	return u, nil
}
