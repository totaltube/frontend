package api

import "sersh.com/totaltube/frontend/types"

func Dmca(siteConfig *types.Config, params types.DmcaParams) (err error) {
	_, err = Request(siteConfig, methodPost, uriDmca, params)
	return
}
