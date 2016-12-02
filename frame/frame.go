package frame

type Frame struct {
	Command string
	Header  Header
	Body    []byte
}

func New() *Frame {
	return &Frame{
		Header: *NewHeader(),
	}
}

func (f *Frame) Clone() *Frame {
	res := Frame{
		Command: f.Command,
		Header:  f.Header,
		Body:    f.Body,
	}
	return &res
}
