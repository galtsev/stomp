package frame

type ParsingError struct {
	msg string
}

func (err ParsingError) Error() string {
	return err.msg
}
