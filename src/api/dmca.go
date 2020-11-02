package api

import "sersh.com/totaltube/frontend/types"

func Dmca(params types.DmcaParams) (err error) {
	_, err = apiRequest(methodPost, uriDmca, params)
	return
}
