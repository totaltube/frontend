package handlers

func InitBackgrounds() {
	// Init some background goroutines
	go doCount()
	go doCount()
	go doCount()
	go doCount()
}
