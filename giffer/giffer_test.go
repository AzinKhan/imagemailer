package giffer

import (
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"path"
	"testing"
)

const imageDir string = "./test_images"

func loadJPG(filename string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	return b, err
}

func TestGiffer(t *testing.T) {
	images, err := ioutil.ReadDir(imageDir)
	if err != nil {
		t.Fail()
	}
	inputData := make([][]byte, len(images))
	for i, img := range images {
		fullPath := path.Join(imageDir, img.Name())
		data, err := loadJPG(fullPath)
		if err != nil {
			log.Printf("Could not load %+v", fullPath)
			t.Fail()
		}
		inputData[i] = data
	}
	// Make GIF
	GIF, err := Giffer(inputData)
	if err != nil {
		log.Printf("Non-nil error: %+v", err)
		t.Fail()
	}
	if GIF.Len() == 0 {
		t.Fail()
	}
}

func TestGifferReturnsError(t *testing.T) {
	d := [][]byte{[]byte("This is not a jpeg image")}
	buf, err := Giffer(d)
	if err == nil {
		t.Fail()
	}
	if buf != nil {
		t.Fail()
	}

}

func TestDecodeJPGError(t *testing.T) {
	// Make random bytes
	d := []byte("This is not a jpeg image")
	img, err := decodeJPG(d)
	if err == nil {
		t.Fail()
	}
	if img != nil {
		t.Fail()
	}
}

type MockImage struct {
	Pix string
}

func (m MockImage) ColorModel() color.Model {
	log.Println("ColorModel")
	var c color.Model
	return c
}

func (m MockImage) Bounds() image.Rectangle {
	min := image.Point{
		X: 0,
		Y: 0,
	}
	max := image.Point{
		X: 1 << 18,
		Y: 1 << 18,
	}
	r := image.Rectangle{
		Min: min,
		Max: max,
	}
	return r

}

func (m MockImage) At(x int, y int) color.Color {
	log.Println("At")
	var c color.Color
	return c
}

func TestConvertToGif(t *testing.T) {
	// Send in a fake large image
	m := MockImage{
		Pix: "hello",
	}
	img, err := ConvertToGIF(m)
	if err == nil {
		log.Println("Nil error")
		t.Fail()
	}
	if img != nil {
		log.Println("Image is not nil")
		t.Fail()
	}

}
