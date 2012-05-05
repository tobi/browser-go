package main

import (
	"flag"
	"github.com/tobi/mogrify-go"
	"io/ioutil"
	"log"
	"fmt"
	"net/http"
	"time"
)

var phantom *Phantom = NewWebkitPool(1)

var port *int = flag.Int("port", 3000, "port")
var cacheLength *int = flag.Int("secs", 3*60, "cache retention in seconds")

func param(r *http.Request, name string) string {
	if len(r.Form[name]) > 0 {
		return r.Form[name][0]
	}
	return ""
}

func serveFile(w http.ResponseWriter, r *http.Request, filename string) {
	http.ServeFile(w, r, filename)
}

// servePng takes a byte slice and flushes it on the wire
// treating it as an image/png mime type
func (p *Process) ServePng(b []byte) {
	// mime
	p.writer.Header().Set("Content-Type", "image/png")

	// 3 hours
	p.writer.Header().Set("Cache-Control", "public, max-age=10800")
	p.writer.WriteHeader(http.StatusOK)
	p.writer.Write(b)
	p.bytesWritten = len(b)	
	p.status = http.StatusOK
}

// Return true if the cache entry should be considered
// fresh based on the command line parameters
func fresh(c *cacheEntry) bool {
	elapsed := time.Since(c.stat.ModTime()).Minutes()
	return elapsed < float64(*cacheLength) * 3
}

func (p *Process) ServeError(msg string) {
	log.Println(msg)
	p.status = http.StatusInternalServerError
	http.Error(p.writer, msg, http.StatusInternalServerError)
}

type Process struct {	
	writer http.ResponseWriter
	request *http.Request

	screenshotUrl string
	screenshotSize string 

	cacheHit bool
	cacheHitRaw bool
	bytesWritten int

	status int
}

func (p *Process) Log() {
	log.Printf("GET %s %dbytes %t %t", p.request.URL.Path, p.status, p.bytesWritten, p.cacheHit, p.cacheHitRaw)
}

func (p *Process) Handle() {
	defer p.Log()
	var buffer []byte
	var cache *cacheEntry

	// Let's just see if we may have a cache hit
	// for the exact p.screenshotUrl + p.screenshotSize
	cache = CacheLookup(p.screenshotUrl + p.screenshotSize)
	if cache != nil && fresh(cache) {

		png, err := ioutil.ReadFile(cache.filepath)

		// can't read the file? 
		if err == nil {
			p.cacheHit = true
			p.ServePng(png)
			return
		}
	}

	// look in the cache for this screenshot
	cache = CacheLookup(p.screenshotUrl)
	if cache != nil && fresh(cache) {
		png, err := ioutil.ReadFile(cache.filepath)
		if err == nil {
			p.cacheHitRaw = true
			buffer = png
		}
	}

	if p.cacheHitRaw == false {

		// make the screenshot
		filename := phantom.Screenshot(p.screenshotUrl)

		if filename == "" {
			p.ServeError("Error creating screenshot")
			return
		}

		png, err := ioutil.ReadFile(filename)
		if err == nil {
			buffer = png
			CacheStore(p.screenshotUrl, buffer)
		}
	}

	if p.screenshotSize == "" {
		p.ServePng(buffer)
		return
	}

	image := mogrify.NewImage()
	defer image.Destroy()

	err := image.OpenBlob(buffer)

	if err != nil {
		p.ServeError(err.Error())
		return
	}

	resized, err := image.NewTransformation("", p.screenshotSize)
	defer resized.Destroy()

	if err != nil {
		p.ServeError(err.Error())
		return
	}

	blob, err := resized.Blob()

	if err != nil {
		p.ServeError(err.Error())
		return
	}

	CacheStore(p.screenshotUrl+p.screenshotSize, blob)
	p.ServePng(blob)
}


func Server(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	process := Process{writer: w, request: r}

	if len(r.Form["src"]) > 0 {
		process.screenshotUrl = r.Form["src"][0]
	} else {
		http.NotFound(w, r)
		return 
	}

	if len(r.Form["size"]) > 0 {
		process.screenshotSize = r.Form["size"][0]
	}

	process.Handle()	
	return
}

func main() {
	flag.Parse()
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/", Server)

	binding := fmt.Sprintf(":%d", *port)
	log.Printf("Running and listening to port %d", *port)

	if err := http.ListenAndServe(binding, nil); err != nil {
		log.Panicln("Could not start server:", err)
	}
}
