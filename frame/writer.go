package frame

import (
	"bufio"
	"io"
)

type FrameWriter interface {
	Write(p []byte) (int, error)
}

type Writer struct {
	w   *bufio.Writer
	buf []byte
}

func NewWriter(w io.Writer) *Writer {
	wr := &Writer{
		w:   bufio.NewWriter(w),
		buf: make([]byte, 240),
	}
	return wr
}

func (w *Writer) Write2(fr *Frame) error {
	wb := w.w.WriteByte
	wr := w.w.Write
	_, err := wr([]byte(fr.Command))
	if err != nil {
		return err
	}
	wb('\n')
	fr.Header.Write(&w.buf)
	wr(w.buf)
	if err != nil {
		return err
	}
	wb('\n')
	_, err = wr(fr.Body)
	if err != nil {
		return err
	}
	wb('\x00')
	w.w.Flush()
	return nil
}

func (w *Writer) Write(fr *Frame) error {
	b := w.buf[:0]
	b = append(b, []byte(fr.Command)...)
	b = append(b, '\n')
	fr.Header.Write(&b)
	b = append(b, '\n')
	_, err := w.w.Write(b)
	if err != nil {
		return err
	}
	_, err = w.w.Write(fr.Body)
	if err != nil {
		return err
	}
	w.w.Write([]byte{0})
	w.w.Flush()
	w.buf = b
	return nil
}
