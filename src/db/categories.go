package db

import (
	"github.com/segmentio/encoding/json"
	"log"
	"math/rand"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
	"sync"
	"time"
)
var topCategoriesCache sync.Map
var topCategoriesCacheExpire sync.Map

// GetCachedTopCategories triple cache for top categories
func GetCachedTopCategories(siteDomain string) (results *types.CategoryResults, err error) {
	lang := "en"
	var cacheKey = "in:topcat:" + siteDomain + ":" + lang
	helpers.KeyMutex.Lock(cacheKey)
	defer helpers.KeyMutex.Unlock(cacheKey)
	if value, ok := topCategoriesCacheExpire.Load(cacheKey); ok && value.(time.Time).After(time.Now()) {
		if v, ok := topCategoriesCache.Load(cacheKey); ok {
			results = v.(*types.CategoryResults)
			return
		}
	}
	var ttl = time.Hour*2+time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = GetCachedTimeout(cacheKey, ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		_, rawResponse, err = api.CategoriesList(siteDomain, lang, 1, api.SortPopular, 150)
		return rawResponse, err
	}, false); err != nil {
		log.Println(err)
		return
	}
	results = new(types.CategoryResults)
	err = json.Unmarshal(cached, results)
	if err == nil {
		topCategoriesCache.Store(cacheKey, results)
		topCategoriesCacheExpire.Store(cacheKey, time.Now().Add(time.Minute*20+time.Duration(rand.Intn(600))*time.Second))
	}
	return
}
