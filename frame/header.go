package frame

import (
	"bytes"
)

const (
	HdrAcceptVersion = "accept-version"
	HdrAck           = "ack" // SUBSCRIBE, MESSAGE
	HdrContentLength = "content-length"
	HdrContentType   = "content-type"
	HdrDestination   = "destination" // SEND, SUBSCRIBE, MESSAGE
	HdrHeartBeat     = "heart-beat"
	HdrHost          = "host"
	HdrId            = "id" // SUBSCRIBE, UNSUBSCRIBE (subscription-id), ACK, NACK (=ack of MESSAGE)
	HdrLogin         = "login"
	HdrMessage       = "message"
	HdrMessageId     = "message-id" // MESSAGE
	HdrPasscode      = "passcode"
	HdrReceipt       = "receipt"
	HdrReceiptId     = "receipt-id"
	HdrServer        = "server"
	HdrSession       = "session"
	HdrSubscription  = "subscription" // MESSAGE
	HdrTransaction   = "transaction"
	HdrVersion       = "version"
)

func Encode(value string, dest *[]byte) {
	*dest = (*dest)[:0]
	for _, c := range []byte(value) {
		switch c {
		case '\\':
			*dest = append(*dest, '\\', '\\')
		case '\r':
			*dest = append(*dest, '\\', 'r')
		case '\n':
			*dest = append(*dest, '\\', 'n')
		case ':':
			*dest = append(*dest, '\\', 'c')
		default:
			*dest = append(*dest, c)
		}
	}
}

func Decode(value []byte) string {
	dest := make([]byte, 0, len(value)+8)
	i := 0
	for i < len(value)-1 {
		c := value[i]
		if c == '\\' {
			switch value[i+1] {
			case '\\':
				dest = append(dest, '\\')
			case 'r':
				dest = append(dest, '\r')
			case 'n':
				dest = append(dest, '\n')
			case 'c':
				dest = append(dest, 'c')
			default:
				panic(ParsingError)
			}
			i += 2
		} else {
			dest = append(dest, c)
			i += 1
		}
	}
	dest = append(dest, value[i])
	return string(dest)
}

type Header struct {
	headers map[string]string
}

func NewHeader() *Header {
	return &Header{
		headers: make(map[string]string, 8),
	}
}

func (h *Header) Get(name string) (value string, ok bool) {
	value, ok = h.headers[name]
	return
}

func (h *Header) Set(name, value string) {
	h.headers[name] = value
}

func (h *Header) Parse(buf []byte) error {
	p := bytes.IndexByte(buf, ':')
	if p < 0 {
		return ParsingError
	}
	h.Set(string(buf[:p]), Decode(buf[p+1:]))
	return nil
}
