package homebrew

import (
	"bytes"
	"testing"
)

type idxReaderAt struct{}

func (f idxReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	for idx := range p {
		p[idx] = byte(off)
		off++
	}
	return len(p), nil
}

func TestOvlReader(t *testing.T) {
	o := &ovlReaderAt{
		f:    idxReaderAt{},
		data: []byte{0xF0, 0xF1, 0xF2, 0xF3},
		off:  0x10,
	}

	var tests = []struct {
		off  int64
		data []byte
	}{
		{0x0, []byte{0, 1, 2, 3, 4, 5}},
		{0xC, []byte{0xC, 0xD, 0xE, 0xF, 0xF0, 0xF1, 0xF2, 0xF3, 0x14, 0x15, 0x16}},

		{0xF, []byte{0xF}},
		{0xF, []byte{0xF, 0xF0}},
		{0xF, []byte{0xF, 0xF0, 0xF1}},
		{0xF, []byte{0xF, 0xF0, 0xF1, 0xF2, 0xF3}},
		{0xF, []byte{0xF, 0xF0, 0xF1, 0xF2, 0xF3, 0x14}},
		{0xF, []byte{0xF, 0xF0, 0xF1, 0xF2, 0xF3, 0x14, 0x15, 0x16}},

		{0x10, []byte{0xF0}},
		{0x10, []byte{0xF0, 0xF1}},
		{0x10, []byte{0xF0, 0xF1, 0xF2, 0xF3}},
		{0x10, []byte{0xF0, 0xF1, 0xF2, 0xF3, 0x14}},
		{0x10, []byte{0xF0, 0xF1, 0xF2, 0xF3, 0x14, 0x15, 0x16}},

		{0x11, []byte{0xF1}},
		{0x11, []byte{0xF1, 0xF2, 0xF3}},
		{0x11, []byte{0xF1, 0xF2, 0xF3, 0x14}},
		{0x11, []byte{0xF1, 0xF2, 0xF3, 0x14, 0x15, 0x16}},

		{0x13, []byte{0xF3}},
		{0x13, []byte{0xF3, 0x14}},
		{0x13, []byte{0xF3, 0x14, 0x15, 0x16}},

		{0x14, []byte{0x14}},
		{0x14, []byte{0x14, 0x15, 0x16}},

		{0x20, []byte{0x20, 0x21, 0x22}},
	}

	for _, test := range tests {
		buf := make([]byte, len(test.data))
		o.ReadAt(buf, test.off)
		if !bytes.Equal(buf, test.data) {
			t.Errorf("invalid read at %x: got=%x, want=%x", test.off, buf, test.data)
		}
	}
}
