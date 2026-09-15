package main

import (
	"bytes"
	_ "embed"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	//go:embed asset.page.js
	assetPageJS []byte

	//go:embed asset.style.css
	assetStyle []byte

	regexpURL = regexp.MustCompile(`https://\S+`)
)

func init() {
	assetPageJS = bytes.ReplaceAll(assetPageJS, []byte{'\n'}, nil)
	assetPageJS = bytes.ReplaceAll(assetPageJS, []byte{'\t'}, nil)
	assetPageJS = bytes.ReplaceAll(assetPageJS, []byte{' '}, nil)

	assetStyle = bytes.ReplaceAll(assetStyle, []byte{'\n'}, nil)
	assetStyle = bytes.ReplaceAll(assetStyle, []byte{'\t'}, nil)
	assetStyle = bytes.ReplaceAll(assetStyle, []byte(" {"), []byte("{"))
	assetStyle = bytes.ReplaceAll(assetStyle, []byte(": "), []byte(":"))
	assetStyle = bytes.ReplaceAll(assetStyle, []byte(";}"), []byte("}"))
}

type CachePage struct {
	LastMod time.Time
	Title   string
	URL     [][2]string
	Content string
}

func (cache *Cache) handPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/m/")
	log.Printf("serve %q", id)
	mainHost, mainId, ok := strings.Cut(id, "@")
	if !ok {
		http.Error(w, "expected: '/m/host@pad_id'", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(cache.genPage(mainHost, mainId))
}

func (cache *Cache) genPage(website, id string) (buff []byte) {
	main := cache.GetPage(website, id)
	if main == nil {
		return genNotReady()
	}
	titleSafe := html.EscapeString(main.Title)

	buff = append(buff, `<!DOCTYPE html><html lang=en>`...)
	buff = append(buff, `<head>`...)
	buff = append(buff, `<meta charset=utf-8>`...)
	buff = append(buff, `<meta name=viewport content="width=device-width,initial-scale=1">`...)
	buff = append(buff, `<link rel=icon href=/favicon.ico type=image/webp>`...)
	buff = append(buff, `<title>`...)
	buff = append(buff, titleSafe...)
	buff = append(buff, `[index] </title>`...)
	buff = append(buff, `<style>`...)
	buff = append(buff, assetStyle...)
	buff = append(buff, `</style>`...)
	buff = append(buff, `</head><body>`...)

	buff = append(buff, `<nav><b>`...)
	buff = appendItem(buff, [2]string{website, id}, main)
	buff = append(buff, `</b>`...)
	if main != nil && len(main.URL) > 0 {
		for _, u := range main.URL {
			buff = appendItem(buff, u, cache.GetPage(u[0], u[1]))
		}
	}

	buff = append(buff, `<hr><a href='`...)
	buff = append(buff, html.EscapeString(about)...)
	buff = append(buff, `'>[about]</a>`...)
	buff = append(buff, `</nav>`...)
	buff = append(buff, `<iframe></iframe>`...)

	buff = append(buff, `<script>`...)
	buff = append(buff, assetPageJS...)
	buff = append(buff, `</script>`...)

	return
}

func genNotReady() (buff []byte) {
	buff = append(buff, `<!DOCTYPE html><html lang=en>`...)
	buff = append(buff, `<head>`...)
	buff = append(buff, `<meta charset=utf-8>`...)
	buff = append(buff, `<meta name=viewport content="width=device-width,initial-scale=1">`...)
	buff = append(buff, `<link rel=icon href=/favicon.ico type=image/webp>`...)
	buff = append(buff, `<meta http-equiv=refresh content=1>`...)
	buff = append(buff, `<style>`...)
	buff = append(buff, `body{font-family:sans-serif;font-size:50px;background:#eb5e28}`...)
	buff = append(buff, `</style>`...)
	buff = append(buff, `</head><body>`...)
	buff = append(buff, `Not ready yet. Please wait and retry in some time...<br>`...)
	return
}

func appendItem(buff []byte, u [2]string, page *CachePage) []byte {
	title := ""
	if page != nil {
		title = page.Title
	} else {
		title = u[0] + "@" + u[1]
	}

	buff = append(buff, `<div class=item data-url='https://`...)
	buff = append(buff, html.EscapeString(u[0])...)
	buff = append(buff, "/p/"...)
	buff = append(buff, html.EscapeString(u[1])...)
	buff = append(buff, `'>* `...)
	buff = append(buff, html.EscapeString(title)...)
	buff = append(buff, `</div>`...)
	return buff
}

func (cache *Cache) GetPage(host, id string) (page *CachePage) {
	cache.Lock()
	defer cache.Unlock()
	page = cache.M[host+"@"+id]
	if page == nil {
		go func() {
			page := FetchPage(host, id)
			if page != nil {
				cache.Lock()
				defer cache.Unlock()
				cache.M[host+"@"+id] = page
			}
		}()
	}
	return
}

func FetchPage(host, id string) *CachePage {
	response, err := http.Get("https://" + host + "/p/" + id + "/export/txt")
	if err != nil {
		log.Printf("fetch %q fail:%v", host+"@"+id, err)
		return nil
	} else if response.StatusCode != 200 {
		log.Printf("fetch %q wrong status: %q", host+"@"+id, response.Status)
		return nil
	}
	log.Printf("fetch %q", host+"@"+id)

	data, err := io.ReadAll(response.Body)
	content := string(data)

	title, _, _ := strings.Cut(content, "\n")
	title = strings.TrimFunc(title, func(r rune) bool {
		switch r {
		case ' ', '\t', '#', '=', '*':
			return true
		default:
			return false
		}
	})

	urls := make([][2]string, 0)
	for _, raw := range regexpURL.FindAllString(content, -1) {
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		path, found := strings.CutPrefix(u.Path, "/p/")
		if !found {
			continue
		}
		urls = append(urls, [2]string{u.Host, path})
	}

	return &CachePage{
		LastMod: time.Now(),
		Title:   title,
		URL:     urls,
		Content: content,
	}
}
