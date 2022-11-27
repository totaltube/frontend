package main

import (
	"log"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/internal"
)

func initLanguages() {
	languages, err := api.Languages("")
	if err != nil {
		log.Fatalln("Can't get languages from api:", err)
	}
	internal.InitLanguages(languages)
}

func initCountryGroups() {
	countryGroups, err := api.CountryGroups()
	if err != nil {
		log.Println("Can't get country group info from api:", err)
		return
	}
	internal.InitCountryGroups(countryGroups)
}