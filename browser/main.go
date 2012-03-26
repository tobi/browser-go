package main

import (
	"log"
	"net/http"
	"browser/phantom"	
)

func param(r *http.Request, name string) string {
	if len(r.Form["src"]) > 0 {
		return r.Form["src"][0]
	}
	return ""
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		log.Printf("%v", r.Form)

		url := param(r, "src")

		if url != "" {
			filename, err := phantom.Screenshot(r.Form["src"][0])

			if err != nil {
				log.Printf("Error creating screenshot: %s", err)
				http.Error(w, "Could not create screenshot", http.StatusInternalServerError)
			} else {
				http.ServeFile(w, r, filename)
				return
			}
		}
		http.Error(w, "missing src parameter", http.StatusInternalServerError)

	})

	port := ":3000"
	log.Printf("Running and listening to port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Panicln("Could not start server:", err)
	}
}
