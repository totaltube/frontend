package internal

import (
	"log"
	"net"

	"github.com/samber/lo"

	"sersh.com/totaltube/frontend/geoip"
	"sersh.com/totaltube/frontend/types"
)

var CountryGroups = make([]types.CountryGroup, 0, 10)
var allGroup = types.CountryGroup{
	Name: "_all",
}

func InitCountryGroups(countryGroups []types.CountryGroup) {
	CountryGroups = countryGroups
	var exists bool
	allGroup, exists = lo.Find(countryGroups, func(c types.CountryGroup) bool { return c.Name == "_all" })
	if !exists {
		log.Fatalln("_all group not defined for country groups")
	}
}

func GetCountryGroup(country string) types.CountryGroup {
	for _, c := range CountryGroups {
		if c.Name == country {
			return c
		}
	}
	return allGroup
}

func DetectCountryGroup(ip net.IP) types.CountryGroup {
	countryCode, _ := geoip.Country(ip)
	for _, c := range CountryGroups {
		if lo.Contains(c.Countries, countryCode) {
			return c
		}
	}
	return allGroup
}
