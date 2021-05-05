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
	var cacheKey = "topcat:" + siteDomain + ":" + lang
	helpers.KeyMutex.Lock(cacheKey)
	defer helpers.KeyMutex.Unlock(cacheKey)
	if value, ok := topCategoriesCacheExpire.Load(cacheKey); ok && value.(time.Time).After(time.Now()) {
		if v, ok := topCategoriesCache.Load(cacheKey); ok {
			results = v.(*types.CategoryResults)
			return
		}
	}
	var data = GetCached(cacheKey)
	if data != nil {
		results = new(types.CategoryResults)
		err = json.Unmarshal(data, results)
		if err == nil {
			topCategoriesCache.Store(cacheKey, results)
			expire := time.Minute*30+time.Duration(rand.Intn(600))*time.Second
			topCategoriesCacheExpire.Store(cacheKey, time.Now().Add(expire))
		}
		return
	}
	var rawResponse json.RawMessage
	results, rawResponse, err = api.CategoriesList(siteDomain, lang, 1, api.SortPopular, 150)
	if err != nil {
		log.Println(err)
		return
	}
	expire := time.Minute*120+time.Duration(rand.Intn(3600))*time.Second
	err = PutCached(cacheKey, rawResponse, expire)
	if err != nil {
		log.Println(err)
		return
	}
	topCategoriesCache.Store(cacheKey, results)
	topCategoriesCacheExpire.Store(cacheKey, time.Now().Add(expire))
	return
}
