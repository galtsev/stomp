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
	res := New()
	res.Command = f.Command
	res.Body = f.Body
	res.Header.Update(f.Header)
	return res
}
