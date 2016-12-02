package frame

import (
	"bufio"
	"io"
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(reader),
	}
}

func (r *Reader) Read() (*Frame, error) {
	fr := New()
	// Command
	cmd, err := r.reader.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	fr.Command = string(cmd[:len(cmd)-1])
	// Headers
	for {
		h, err := r.reader.ReadSlice('\n')
		if err != nil {
			return nil, err
		}
		if len(h) == 1 {
			// empty line - end of headers
			break
		}
		err = fr.Header.Parse(h[:len(h)-1])
		if err != nil {
			return nil, err
		}
	}
	tmpBody, err := r.reader.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	fr.Body = tmpBody[:len(tmpBody)-1] // strip terminating zero byte
	return fr, nil
}
