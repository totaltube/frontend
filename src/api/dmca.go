package api

import "sersh.com/totaltube/frontend/types"

func Dmca(siteDomain string, params types.DmcaParams) (err error) {
	_, err = apiRequest(siteDomain, methodPost, uriDmca, params)
	return
}
