package api

import (
	"encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

type CommentsSortBy string

const (
	CommentsSortByLastUpdated CommentsSortBy = "updated"
	CommentsSortByCreated     CommentsSortBy = "created"
	CommentsSortByLikes       CommentsSortBy = "likes"
	CommentsSortByDislikes    CommentsSortBy = "dislikes"
)

func GetComments(
	siteConfig *types.Config, contentId int64, from int, size int, sort CommentsSortBy, lang string,
) (results *types.CommentsResult, response json.RawMessage, err error) {
	response, err = Request(siteConfig, methodGet, uriCommentsGet, url.Values{
		"lang":       []string{lang},
		"sort":       []string{string(sort)},
		"size":       []string{strconv.FormatInt(int64(size), 10)},
		"from":       []string{strconv.FormatInt(int64(from), 10)},
		"content_id": []string{strconv.FormatInt(contentId, 10)},
	})
	if err != nil {
		return
	}
	results = &types.CommentsResult{}
	err = json.Unmarshal(response, results)
	return
}

func GetReplyComments(siteConfig *types.Config, commentId int64) (results *types.ReplyCommentsResult, response json.RawMessage, err error) {
	response, err = Request(siteConfig, methodGet, uriCommentsReplies, url.Values{
		"comment_id": []string{strconv.FormatInt(commentId, 10)},
	})
	if err != nil {
		return
	}
	results = &types.ReplyCommentsResult{}
	err = json.Unmarshal(response, results)
	return
}