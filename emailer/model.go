package emailer

import "bytes"

type Email struct {
	Subject    string
	Body       []byte
	Attachment Attachment
}

type Sender interface {
	Send(e Email) error
}

type Processor interface {
	Process(upload []byte)
}

type Attachment struct {
	Data        *bytes.Buffer
	Filename    string
	ContentType string
}
