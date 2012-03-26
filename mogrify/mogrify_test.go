package mogrify

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"
)

func copyFile(src, dst string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

func TestResizeFile(t *testing.T) {

	targetPath := os.TempDir() + "/image1.jpg"
	_, err := copyFile("../assets/image.jpg", targetPath)

	if err != nil {		
		log.Printf("Could not copy file to tmp folder: %s", err)
		t.Fail()
	}

	ResizeFile(targetPath, "50x50")
}

func TestResizeStream(t *testing.T) {
	image, err := os.Open("../assets/image.jpg")
	if err != nil {
		t.Fail()
	}

	output, err := os.Create( os.TempDir() + "/image2.jpg")
	if err != nil {
		t.Fail()
	}

	ResizeStream(output, image, "50x50")
}

func TestResizeFileDoesntExist(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	in := bytes.NewBuffer(nil)
	err := Resize(buf, in, "50x50")
	if err == nil {
		t.Fail()
	}
}
