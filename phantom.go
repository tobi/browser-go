package main

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

type shot struct {
	url        string
	resultChan chan string
}

func init() {
	dir = os.TempDir()
	rand = reseed()

	var err error

	phantomPath, err = exec.LookPath("phantomjs")

	if err != nil {
		log.Fatalf("Cannot find phantomjs executable in bath")
	}

}

type Phantom struct {
	screenshotChan chan *shot
}

func NewWebkitPool(pool int) *Phantom {
	phantom := &Phantom{make(chan *shot, 10)}

	for i := 0; i < pool; i++ {
		go phantom.webkitWorker()
	}

	return phantom
}

func (p *Phantom) webkitWorker() {
	for shot := range p.screenshotChan {

		result, err := screenshot(shot.url)
		if err == nil {
			shot.resultChan <- result
		} else {
			log.Printf("Screenshot error: %s", err)
			close(shot.resultChan)
		}

	}
}

func (p *Phantom) Screenshot(url string) string {
	shot := shot{url, make(chan string)}
	p.screenshotChan <- &shot
	return <-shot.resultChan
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

func screenshot(url string) (string, error) {
	log.Printf("Screenshotting %s", url)

	filename := tempPngFileName()

	cmd := exec.Command(phantomPath, "render.js", url, "1280", "500", filename)

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
