package main

import (
	_ "embed"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	//go:embed favicon.webp
	faviconWEBP []byte

	//go:embed home.html
	homeHTML []byte
)

func main() {
	cache := &Cache{
		M: make(map[string]*CachePage),
	}
	go func() {
		for range time.Tick(5 * time.Second) {
			cache.Clean()
		}
	}()
	log.Println("listen ...")
	log.Fatal(http.ListenAndServe(":8000", cache))
}

type Cache struct {
	sync.RWMutex
	M map[string]*CachePage
}

func (cache *Cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		cache.handPage(w, r)
	}
}

func (cache *Cache) Clean() {
	cache.Lock()
	defer cache.Unlock()
	now := time.Now()
	for k, v := range cache.M {
		if v == nil {
			delete(cache.M, k)
		} else if now.Sub(v.Modif) > 5*time.Minute {
			delete(cache.M, k)
		}
	}
}
