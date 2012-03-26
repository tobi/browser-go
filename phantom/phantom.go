package phantom

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

var phantomPath string
var rand uint32
var dir string

func init() {
	dir = os.TempDir()
	rand = reseed()

	var err error

	phantomPath, err = exec.LookPath("phantomjs")

	if err != nil {
		log.Fatalf("Cannot find phantomjs executable in bath")
	}
}

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() string {
	r := rand
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

func tempPngFileName() string {
	return filepath.Join(dir, "img", nextSuffix(), ".png")
}

func Screenshot(url string) (string, error) {
	log.Printf("Screenshotting %s", url)

	filename := tempPngFileName()

	cmd := exec.Command(phantomPath, "render.js", url, "1024", "786", filename)

	// connect to STDIN / STDOUT
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Printf("Error: %s", err)
		return "", err
	}

	return filename, nil
}
