package helpers

import "github.com/samber/lo"

func FirstNotEmpty[T comparable](args ...T) (result T) {
	result, _ = lo.Coalesce(args...)
	return
}
