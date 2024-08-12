package types

import "time"

// Comment - top level comment to content
type Comment struct {
	CommentId int64          `json:"comment_id"`
	Text      string         `json:"text"`
	Created   time.Time      `json:"created"`
	Updated   time.Time      `json:"updated"`
	Likes     int32          `json:"likes,omitempty"`
	Dislikes  int32          `json:"dislikes,omitempty"`
	UserId    int32          `json:"user_id"`
	Username  string         `json:"username"`
	UserIp    string         `json:"user_ip"`
	Language  string         `json:"language,omitempty"`
	Replies   []ReplyComment `json:"replies"`
}

// ReplyComment - reply to comment
type ReplyComment struct {
	CommentId        int64     `json:"comment_id"`
	ReplyToCommentId int64     `json:"reply_to_comment_id"`
	ReplyToUserId    int32     `json:"reply_to_user_id"`
	ReplyToUsername  string    `json:"reply_to_username"`
	UserId           int32     `json:"user_id"`
	Username         string    `json:"username"`
	UserIp           string    `json:"user_ip"`
	Text             string    `json:"text"`
	Created          time.Time `json:"created"`
	Likes            int32     `json:"likes,omitempty"`
	Dislikes         int32     `json:"dislikes,omitempty"`
	Language         string    `json:"language,omitempty"`
}

type CommentsResult struct {
	Items []Comment `json:"items"`
	Total    int64     `json:"total"`
}
type ReplyCommentsResult struct {
	Items []ReplyComment `json:"items"`
}
