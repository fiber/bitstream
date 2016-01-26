package bitstream

import "io"

type (
	// Bit represents a bit 0=false;1=true
	Bit bool
	// Writer writes bits to an underlying writer
	Writer struct {
		w   io.Writer
		b   byte
		p   uint16
		bw  io.ByteWriter
		buf []byte
		err error
	}
)

const (
	defaultBufferSize = 128
)

// constant Bit values
const (
	Zero Bit = false
	One  Bit = true
	Z        = Zero
	O        = One
)

// NewByteWriter creates a new writer, writing directly to an underlying byte writer
func NewByteWriter(w io.ByteWriter) *Writer {
	return &Writer{bw: w, p: 8}
}

// NewWriter creates a new buffered writer, using a default buffer size (128 bytes)
func NewWriter(w io.Writer) *Writer {
	return NewWriterSize(w, defaultBufferSize)
}

// NewWriterSize writes to an underlying buffer
func NewWriterSize(w io.Writer, n int) *Writer {
	return NewWriterBuffer(w, make([]byte, 0, n))
}

// NewWriterBuffer creates a writer that writes to the underlying writer through
// the provided buffer
func NewWriterBuffer(w io.Writer, b []byte) *Writer {
	return &Writer{w: w, p: 8, buf: b}
}

func (w *Writer) push(b byte) error {
	if w.bw != nil {
		if err := w.bw.WriteByte(b); err != nil {
			w.err = err
		}
		return w.err
	}
	w.buf = append(w.buf, b)
	if len(w.buf) < cap(w.buf) {
		return w.err
	}
	if _, err := w.w.Write(w.buf); err != nil {
		w.err = err
	}
	w.buf = w.buf[:0]
	return w.err
}

// WriteZero writes a Zero bit into the stream
func (w *Writer) WriteZero() error {
	if c := w.p; c > 1 {
		w.p = c - 1
		return w.err
	}
	w.push(w.b)
	w.b = 0
	w.p = 8
	return w.err
}

// WriteOne writes a one bit into the stream
func (w *Writer) WriteOne() error {
	if c := w.p; c > 1 {
		c--
		w.b |= 1 << c
		w.p = c
		return w.err
	}
	w.push(w.b | 1)
	w.b = 0
	w.p = 8
	return w.err
}

// WriteBit writes the given bit to the stream
func (w *Writer) WriteBit(bit Bit) error {
	if bit {
		return w.WriteOne()
	}
	return w.WriteZero()
}

// WriteByte writes a single byte to the stream
func (w *Writer) WriteByte(b byte) error {
	w.push(w.b | b>>(8-w.p))
	w.b = b << w.p
	return w.err
}

// Write writes a byte array into the stream
func (w *Writer) Write(b []byte) error {
	if w.p == 8 {
		if w.bw != nil {
			for _, n := range b {
				if err := w.bw.WriteByte(n); err != nil {
					w.err = err
				}
			}
		}
		if len(w.buf) > 0 {
			if _, err := w.w.Write(w.buf); err != nil {
				w.err = err
			}
		}
		if _, err := w.w.Write(b); err != nil {
			w.err = err
		}
	} else {
		wb := w.b
		wp := w.p
		wp8 := 8 - w.p
		for _, n := range b {
			w.push(wb | n>>wp8)
			wb = n << wp
		}
		w.b = wb
	}
	return w.err
}

// WriteBits writes the n least significant bits of u, most-significant bit first
func (w *Writer) WriteBits(u uint64, n uint) error {
	u <<= (64 - n)
	for n >= 8 {
		w.WriteByte(byte(u >> 56))
		u <<= 8
		n -= 8
	}
	for n > 0 {
		// Todo(se): this can be 1-2 masks plus max one push.
		if (u >> 63) == 0 {
			w.WriteZero()
		} else {
			w.WriteOne()
		}
		u <<= 1
		n--
	}
	return w.err
}

// Flush flushes any unwritten bits to the underlying writer, potentially
// filling up the last byte with zeros
func (w *Writer) Flush() error {
	for w.p != 8 {
		w.WriteZero()
	}
	return w.flush()
}

// FlushOnes flushes any unwritten bits to the underlying writer, potentially
// filling up the last byte with ones
func (w *Writer) FlushOnes() error {
	for w.p != 8 {
		w.WriteOne()
	}
	return w.flush()
}

func (w *Writer) flush() error {
	if len(w.buf) > 0 {
		if _, err := w.w.Write(w.buf); err != nil {
			w.err = err
		}
	}
	return w.err
}
