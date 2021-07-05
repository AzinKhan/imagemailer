package emailer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AzinKhan/giffer"
)

const (
	timeout           = 20 * time.Second
	maxAttachmentSize = 2000000
)

type ImageProcessor struct {
	sender Sender
	imch   chan []byte
	outch  chan Attachment
}

func NewImageProcessor(ctx context.Context, sender Sender) *ImageProcessor {
	p := &ImageProcessor{
		sender: sender,
		imch:   make(chan []byte),
		outch:  make(chan Attachment),
	}

	p.run(ctx)

	return p
}

func (p *ImageProcessor) Process(data []byte) {
	p.imch <- data
}

// run starts the ImageProcessor, ready to receive uploads
// via the Process method.
func (p *ImageProcessor) run(ctx context.Context) {

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case attachment, ok := <-p.outch:
				if !ok {
					return
				}

				body := fmt.Sprintf("Email sent: %+v", time.Now())
				msg := Email{
					Subject:    fmt.Sprintf("Motion detected! Time: %+v", time.Now()),
					Body:       []byte(body),
					Attachment: attachment,
				}
				err := p.sender.Send(msg)
				if err != nil {
					log.Printf("Error sending email: %v", err)
				}
			}
		}
	}()

	go p.buffer(ctx)

}

func (p *ImageProcessor) buffer(ctx context.Context) {
	t := time.NewTimer(timeout)

	var data [][]byte
	size := 0

	produceGIF := func() {
		GIF, err := makeGIFAttachment(data)
		if err != nil {
			log.Println(err)
			return
		}
		p.outch <- GIF
		size = 0
		data = nil
		t.Reset(timeout)
	}
	// Either append until memory limit reached or
	// until timeout
	for {
		select {
		case <-ctx.Done():
			close(p.outch)
			return
		case <-t.C:
			if size == 0 {
				t.Reset(timeout)
				continue
			}
			produceGIF()
		case d := <-p.imch:
			t.Reset(timeout)
			data = append(data, d)
			size += len(d)
			if size >= maxAttachmentSize {
				produceGIF()
			}
		}
	}
}

func makeGIFAttachment(data [][]byte) (Attachment, error) {
	log.Printf("Combining %d images into GIF attachment\n", len(data))
	GIF, err := giffer.Giffer(data)
	if err != nil {
		return Attachment{}, err
	}
	a := Attachment{
		Data:        GIF,
		Filename:    fmt.Sprintf("%+v.gif", time.Now()),
		ContentType: "image/gif",
	}
	return a, nil
}
