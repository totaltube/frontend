package db

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

var topCategoriesCache sync.Map
var topCategoriesCacheExpire sync.Map

// GetCachedTopCategories triple cache for top categories
func GetCachedTopCategories(siteConfig *types.Config, requestHost string, groupID int64) (results *types.CategoryResults, err error) {
	lang := "en"
	var cacheKey = "in:topcat:" + requestHost + ":" + lang + ":" + strconv.FormatInt(groupID, 10)
	helpers.KeyMutex.Lock(cacheKey)
	defer helpers.KeyMutex.Unlock(cacheKey)
	if value, ok := topCategoriesCacheExpire.Load(cacheKey); ok && value.(time.Time).After(time.Now()) {
		if v, ok := topCategoriesCache.Load(cacheKey); ok {
			results = v.(*types.CategoryResults)
			return
		}
	}
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = GetCachedTimeout(cacheKey, ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		_, rawResponse, err = api.CategoriesList(siteConfig, lang, 1, api.SortPopular, 150, groupID)
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
