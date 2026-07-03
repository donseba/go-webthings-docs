package website

import (
	"context"
	"math/rand"
	"net/http"

	partial "github.com/donseba/go-partial"
)

func renderStandalonePage(w http.ResponseWriter, req *http.Request, status int, data PageData, page *partial.Partial) {
	data = pageDefaults(req, data)
	data.SEO = pageSEO(data)
	page.SetDot(data)
	out, err := partial.RenderWithRequest(context.Background(), req, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(out))
}

func pageDefaults(req *http.Request, data PageData) PageData {
	reqHost := data.Host
	if reqHost == "" && req != nil {
		reqHost = req.Host
	}
	if reqHost == "" {
		reqHost = "gowebthings.com"
	}
	linkReq := &http.Request{Host: reqHost}
	if data.MainURL == "" {
		data.MainURL = mainWebsiteURL(linkReq)
	}
	if data.DocsURL == "" {
		data.DocsURL = docsElementURL(linkReq, "go-partial")
	}
	if data.ShowcaseURL == "" {
		data.ShowcaseURL = mainFamilyURL(linkReq, "showcase", "go-partial")
	}
	if data.SourceURL == "" {
		data.SourceURL = "https://github.com/donseba"
	}
	if data.CurrentPath == "" {
		data.CurrentPath = requestPath(req)
	}
	if data.Production == "" {
		data.Production = productionURL(req, data.Section)
	}
	if data.Local == "" {
		data.Local = localURL(req, data.Section)
	}
	if data.Bulletin == "" {
		data.Bulletin = randomBulletin()
	}
	if data.PromptFace == "" {
		data.PromptFace = docsPromptFaces[rand.Intn(len(docsPromptFaces))]
	}
	return data
}
