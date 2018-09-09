package giffer

import (
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
