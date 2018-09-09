package giffer

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"log"
)

func decodeJPG(data []byte) (image.Image, error) {
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, err
}

func Giffer(inputData [][]byte) (*bytes.Buffer, error) {
	G := &gif.GIF{
		LoopCount: 0,
		Disposal:  nil,
		Delay:     make([]int, len(inputData)),
		Image:     make([]*image.Paletted, len(inputData)),
	}
	log.Println("Converting images to GIF")
	for i, data := range inputData {
		img, err := decodeJPG(data)
		if err != nil {
			return nil, err
		}
		var b []byte
		bf := bytes.NewBuffer(b)
		var opt gif.Options
		opt.NumColors = 256
		err = gif.Encode(bf, img, &opt)
		if err != nil {
			return nil, err
		}
		im, err := gif.Decode(bf)
		if err != nil {
			return nil, err
		}
		G.Image[i] = im.(*image.Paletted)
		G.Delay[i] = 8
	}
	log.Printf("Encoding %+v images into GIF", len(G.Image))
	var buf []byte
	Buf := bytes.NewBuffer(buf)
	err := gif.EncodeAll(Buf, G)
	return Buf, err
}