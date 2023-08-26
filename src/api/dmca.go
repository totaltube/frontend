package api

import "sersh.com/totaltube/frontend/types"

func Dmca(siteDomain string, params types.DmcaParams) (err error) {
	_, err = Request(siteDomain, methodPost, uriDmca, params)
	return
}
