package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"strings"

	"server/pkg/app/query"
	"server/pkg/app/service"
)

type Handler interface {
	Index(w http.ResponseWriter, r *http.Request)
	ShortenCreate(w http.ResponseWriter, r *http.Request)
	ShortenResult(w http.ResponseWriter, r *http.Request)
	RedirectLongUrl(w http.ResponseWriter, r *http.Request)
	BatchSave(w http.ResponseWriter, r *http.Request)
}

type Result struct {
	ShortUrl string
}

type pathsSerializable struct {
	Paths map[string]string `json:"paths"`
}

func NewHandler(
	ctx context.Context,
	urlService service.UrlService,
	urlQueryService query.UrlQueryService,
) Handler {
	return &handler{
		ctx:             ctx,
		urlService:      urlService,
		urlQueryService: urlQueryService,
	}
}

type handler struct {
	ctx             context.Context
	urlService      service.UrlService
	urlQueryService query.UrlQueryService
}

func (h *handler) Index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmplParsed, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmplParsed.Execute(w, nil)
	if err != nil {
		log.Panic(err)
	}
}

func (h *handler) ShortenCreate(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.FormValue("short_url")
	longUrl := r.FormValue("long_url")

	shortUrl = urlAfterLastSlash(shortUrl)
	err := h.urlService.Save(h.ctx, shortUrl, longUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirectURL := fmt.Sprintf("/shorten/result?short_url=%s", shortUrl)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *handler) ShortenResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	shortUrl := r.URL.Query().Get("short_url")
	if shortUrl == "" {
		http.Error(w, "short_url parameter is required", http.StatusBadRequest)
		return
	}

	data := Result{
		ShortUrl: shortUrl,
	}

	tmplParsed, err := template.ParseFiles("./templates/summary.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmplParsed.Execute(w, data)
	if err != nil {
		log.Panic(err)
	}
}

func (h *handler) RedirectLongUrl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	url := vars["url"]
	log.Println(url)

	longUrl, err := h.urlQueryService.GetLongUrlByShortUrl(h.ctx, url)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	log.Println(longUrl)
	http.Redirect(w, r, longUrl, http.StatusPermanentRedirect)
}

func (h *handler) BatchSave(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form data", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("import")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			http.Error(w, "File is required", http.StatusBadRequest)
		} else {
			http.Error(w, "Error reading file", http.StatusBadRequest)
		}
		return
	}
	defer file.Close()

	var paths pathsSerializable
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&paths)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(paths.Paths) == 0 {
		http.Error(w, "No paths provided", http.StatusBadRequest)
		return
	}

	for key, value := range paths.Paths {
		key = urlAfterLastSlash(key)
		h.urlService.Save(h.ctx, key, value)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func urlAfterLastSlash(url string) string {
	pos := strings.LastIndex(url, "/")
	if pos == -1 {
		return url
	}
	return url[pos+1:]
}
