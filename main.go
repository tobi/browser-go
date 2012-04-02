package main

import (
	"browser/mogrify"
	"browser/phantom"
	"bytes"
	"flag"
	"time"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func param(r *http.Request, name string) string {
	if len(r.Form[name]) > 0 {
		return r.Form[name][0]
	}
	return ""
}

func serveFile(w http.ResponseWriter, r *http.Request, filename string) {
	http.ServeFile(w, r, filename)
}

func servePng(w http.ResponseWriter, file io.Reader) {
	// mime
	w.Header().Set("Content-Type", "image/png")

	// 3 hours
	w.Header().Set("Cache-Control", "public, max-age=10800")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}

func fresh(c *cacheEntry) bool {
	elapsed := time.Since(c.stat.ModTime()).Minutes()
	log.Printf("Since last mod: %v", elapsed)

	return time.Since(c.stat.ModTime()).Minutes() <  0.5
}

func httpError(w http.ResponseWriter, msg string) {
	log.Print(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}

func Server(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := param(r, "src")
	size := param(r, "size")

	if url == "" {
		http.NotFound(w, r)
		return
	}

	var buffer  []byte
	var cache *cacheEntry

	// Let's just see if we may have a cache hit
	// for the exact url + size
	cache = CacheLookup(url + size)
	if cache != nil && fresh(cache) {
		png, err := ioutil.ReadFile(cache.filepath)
	
		if err == nil {			
			servePng(w, bytes.NewBuffer(png))
			return
		}
	}

	// look in the cache for this screenshot
	cache = CacheLookup(url)
	if cache != nil && fresh(cache) {
		png, err := ioutil.ReadFile(cache.filepath)
		if err == nil {
			buffer = png
		}
	}

	if len(buffer) == 0 {

		// make the screenshot
		filename := phantom.Screenshot(url)

		if filename == "" {
			httpError(w, "Error creating screenshot")
			return
		}

		png, err := ioutil.ReadFile(filename)
		if err == nil {
			buffer = png
			CacheStore(url, buffer)		
		}
	}

	if size == "" {
		servePng(w, bytes.NewBuffer(buffer))
		return
	}

	var output = new(bytes.Buffer)

	if err := mogrify.Resize(output, bytes.NewBuffer(buffer), size); err != nil {
		httpError(w, "could not resize")
		return
	}

	CacheStore(url + size, output.Bytes())
	servePng(w, output)
	return
}

func main() {
	flag.Parse()
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/", Server)

	port := ":3000"
	log.Printf("Running and listening to port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Panicln("Could not start server:", err)
	}
}
