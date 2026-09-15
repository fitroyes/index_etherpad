package main

import (
	_ "embed"
	"log"
	"net/http"
	"strings"
	"sync"
)

var (
	//go:embed favicon.webp
	faviconWEBP []byte

	//go:embed home.html
	homeHTML []byte
)

func main() {
	http.HandleFunc("/m/", handPage)

	log.Println("listen ...")
	log.Fatal(http.ListenAndServe(":8000", &Cache{
		M: make(map[string]*CachePage),
	}))
}

type Cache struct {
	sync.Mutex
	M map[string]*CachePage
}

func (c *Cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		w.Header().Set("Content-Type", "text/html")
		w.Write(homeHTML)
	case r.URL.Path == "/favicon.ico":
		w.Header().Set("Content-Type", "image/webp")
		w.Write(faviconWEBP)
	case r.URL.Path == "/robots.txt":
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\r\nDisallow: /\r\n"))
	case strings.HasPrefix(r.URL.Path, "/m/"):
		handPage(w, r)
	}
}
