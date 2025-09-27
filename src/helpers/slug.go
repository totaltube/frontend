package helpers

import "github.com/gosimple/slug"

func Slugify(s string) string {
	return slug.Make(s)
}
