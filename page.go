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
	Modif   time.Time
	Title   string
	URL     [][2]string
	Content string
}

func (cache *Cache) handPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/m/")
	log.Printf("meta_page %q", id)
	mainWebsite, mainId, ok := strings.Cut(id, "@")
	if !ok {
		http.Error(w, "expected: '/m/host@pad_id'", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(cache.genItem(mainWebsite, mainId))
}

func (cache *Cache) genItem(website, id string) (buff []byte) {
	main := cache.GetPage(website, id)
	titleSafe := "?"
	if main != nil {
		titleSafe = html.EscapeString(main.Title)
	}

	buff = make([]byte, 0)
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
	buff = append(buff, ``...)
	buff = append(buff, ``...)
	buff = append(buff, ``...)

	buff = append(buff, `<script>`...)
	buff = append(buff, assetPageJS...)
	buff = append(buff, `</script>`...)

	return
}

func appendItem(buff []byte, u [2]string, page *CachePage) []byte {
	if page == nil {
		buff = append(buff, `<div class="fail item" data-url="https://`...)
		buff = append(buff, html.EscapeString(u[0])...)
		buff = append(buff, "/p/"...)
		buff = append(buff, html.EscapeString(u[1])...)
		buff = append(buff, `">* (Fail)@ `...)
		buff = append(buff, html.EscapeString(u[0])...)
		buff = append(buff, "@"...)
		buff = append(buff, html.EscapeString(u[1])...)
		buff = append(buff, "</div>"...)
		return buff
	} else {
		buff = append(buff, `<div class=item data-url='https://`...)
		buff = append(buff, html.EscapeString(u[0])...)
		buff = append(buff, "/p/"...)
		buff = append(buff, html.EscapeString(u[1])...)
		buff = append(buff, `'>* `...)
		buff = append(buff, html.EscapeString(page.Title)...)
		buff = append(buff, `</div>`...)
		return buff
	}
}

func (cache *Cache) GetPage(website, id string) *CachePage {
	cache.RLock()
	if page := cache.M[website+"@"+id]; page != nil {
		cache.RUnlock()
		return page
	}
	cache.RUnlock()

	page := FetchPage(website, id)
	cache.Lock()
	cache.M[website+"@"+id] = page
	cache.Unlock()
	return page
}

func FetchPage(website, id string) *CachePage {
	response, err := http.Get("https://" + website + "/p/" + id + "/export/txt")
	if err != nil {
		log.Printf("fetch %q fail:%v", website+"@"+id, err)
		return nil
	} else if response.StatusCode != 200 {
		log.Printf("fetch %q wrong status: %q", website+"@"+id, response.Status)
		return nil
	}
	log.Printf("serve %q", website+"@"+id)

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

	time.Sleep(2 * time.Second)

	return &CachePage{
		Modif:   time.Now(),
		Title:   title,
		URL:     urls,
		Content: content,
	}
}
