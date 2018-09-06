package giffer

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"log"
)

func convertPalette(input *image.YCbCr) *image.Paletted {
	var grey color.Gray
	paletted := image.NewPaletted(input.Rect, []color.Color{grey})
	paletted.Pix = input.Y
	paletted.Stride = input.YStride
	return paletted
}

func Giffer(inputData [][]byte) (*bytes.Buffer, error) {
	G := &gif.GIF{
		LoopCount: 0,
		Disposal:  nil,
		Delay:     make([]int, len(inputData)),
		Image:     make([]*image.Paletted, len(inputData)),
	}
	for i, data := range inputData {
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			log.Println(err)
		}
		log.Println("Converting image to paletted")
		G.Image[i] = convertPalette(img.(*image.YCbCr))
		G.Delay[i] = 3
	}
	log.Printf("Encoding %+v images into GIF", len(G.Image))
	var buf []byte
	Buf := bytes.NewBuffer(buf)
	err := gif.EncodeAll(Buf, G)
	return Buf, err
}
