package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

var GetContentId = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value("config").(*types.Config)
	urlParam := r.URL.Query().Get("url")
	if urlParam == "" {
		render.JSON(w, r, M{
			"success": false,
			"value":   "url not found",
		})
		return
	}

	// Парсим URL для извлечения path
	parsedUrl, err := url.Parse(urlParam)
	if err != nil {
		render.JSON(w, r, M{
			"success": false,
			"value":   "invalid url",
		})
		return
	}

	// Получаем путь без query параметров
	urlPath := parsedUrl.Path

	// Создаем временный router для парсинга content-item роута
	if config.Routes.ContentItem == "" || config.Routes.ContentItem == "-" {
		render.JSON(w, r, M{
			"success": false,
			"value":   "content item route not configured",
		})
		return
	}

	// Подготавливаем роут для chi (заменяем {page} и {id} на regex patterns)
	routePattern := config.Routes.ContentItem
	routePattern = strings.ReplaceAll(routePattern, "{page}", "{page:[0-9]+}")
	routePattern = strings.ReplaceAll(routePattern, "{id}", "{id:[0-9]+}")

	// Создаем временный router для матчинга
	tempRouter := chi.NewRouter()
	var matchedSlug string
	var matchedId int64
	var routeMatched bool

	matchHandler := func(w http.ResponseWriter, r *http.Request) {
		matchedSlug, _ = url.PathUnescape(chi.URLParam(r, "slug"))
		matchedId, _ = strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		routeMatched = true
	}

	// Добавляем основной роут
	tempRouter.Get(routePattern, matchHandler)

	// Если включена мультиязычность, добавляем роуты для каждого языка
	if config.General.MultiLanguage {
		langTemplate := config.Routes.LanguageTemplate
		languages := internal.GetLanguages(config)
		for _, lang := range languages {
			langRoute := strings.ReplaceAll(langTemplate, "{route}", routePattern)
			langRoute = strings.ReplaceAll(langRoute, "{lang}", lang.Id)
			tempRouter.Get(langRoute, matchHandler)
		}
	}

	// Создаем fake request для матчинга
	fakeReq, _ := http.NewRequest("GET", urlPath, nil)
	fakeWriter := &fakeResponseWriter{}

	// Пытаемся проматчить роут
	tempRouter.ServeHTTP(fakeWriter, fakeReq)

	if !routeMatched {
		render.JSON(w, r, M{
			"success": false,
			"value":   "url does not match content item route",
		})
		return
	}

	// Если нет ни id ни slug - ошибка
	if matchedId == 0 && matchedSlug == "" {
		render.JSON(w, r, M{
			"success": false,
			"value":   "no content identifier found in url",
		})
		return
	}
	// Декскориваем ID если нужно
	if matchedId > 0 && config.Routes.IdXorKey > 0 {
		matchedId = matchedId ^ config.Routes.IdXorKey
	}
	// Если есть только slug, но нет ID - получаем ID через API миньона
	if matchedId == 0 && matchedSlug != "" {
		hostName := r.Context().Value("hostName").(string)
		contentIdResult, err := api.ContentIdBySlug(hostName, matchedSlug)
		if err != nil {
			render.JSON(w, r, M{
				"success": false,
				"value":   "failed to get content id by slug: " + err.Error(),
			})
			return
		}
		matchedId = contentIdResult.Id
	}

	render.JSON(w, r, M{
		"success": true,
		"value": M{
			"id":   matchedId,
			"slug": matchedSlug,
		},
	})
})

// fakeResponseWriter для использования с временным router
type fakeResponseWriter struct {
	statusCode int
	headers    http.Header
}

func (f *fakeResponseWriter) Header() http.Header {
	if f.headers == nil {
		f.headers = make(http.Header)
	}
	return f.headers
}

func (f *fakeResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (f *fakeResponseWriter) WriteHeader(statusCode int) {
	f.statusCode = statusCode
}
