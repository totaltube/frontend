package helpers

import (
	"encoding/base64"
	"math/rand"
	"sort"
	"time"
)

func RandStr(len int) string {
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)
	// Base 64 can be longer than len
	return str[:len]
}

// RandomizeItems randomizes items with a given ratio.
func RandomizeItems[T any](items []T, ratio float64) {
	randSource := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSource)
	n := len(items)

	if ratio <= 0 || n == 0 {
		// Nothing to do
		return
	}
	if ratio >= 1 {
		// Shuffle items completely
		r.Shuffle(n, func(i, j int) {
			items[i], items[j] = items[j], items[i]
		})
		return
	}

	type itemWithPriority struct {
		item     T
		priority float64
	}

	itemsWithPriority := make([]itemWithPriority, n)
	for i := 0; i < n; i++ {
		normalizedPosition := float64(n-i) / float64(n)
		randomValue := r.Float64()
		priority := (1-ratio)*normalizedPosition + ratio*randomValue
		itemsWithPriority[i] = itemWithPriority{
			item:     items[i],
			priority: priority,
		}
	}

	sort.SliceStable(itemsWithPriority, func(i, j int) bool {
		return itemsWithPriority[i].priority > itemsWithPriority[j].priority
	})

	for i := 0; i < n; i++ {
		items[i] = itemsWithPriority[i].item
	}
}
