package reference

import (
	"embed"
	"encoding/json"
	"strings"
)

type Province struct {
	Name     string   `json:"province"`
	Counties []string `json:"counties"`
}

//go:embed iran_locations.json
var files embed.FS

var provinces []Province

func init() {
	data, err := files.ReadFile("iran_locations.json")
	if err != nil || json.Unmarshal(data, &provinces) != nil {
		panic("invalid embedded Iran locations data")
	}
}

func Provinces() []Province {
	out := make([]Province, len(provinces))
	for i, province := range provinces {
		out[i] = Province{Name: province.Name, Counties: append([]string(nil), province.Counties...)}
	}
	return out
}

func ProvinceNames() []string {
	out := make([]string, len(provinces))
	for i, province := range provinces {
		out[i] = province.Name
	}
	return out
}

func Counties(provinceName string) ([]string, bool) {
	provinceName = strings.TrimSpace(provinceName)
	for _, province := range provinces {
		if province.Name == provinceName {
			return append([]string(nil), province.Counties...), true
		}
	}
	return nil, false
}

func ValidAddress(province, county string) bool {
	counties, ok := Counties(province)
	if !ok {
		return false
	}
	county = strings.TrimSpace(county)
	for _, item := range counties {
		if item == county {
			return true
		}
	}
	return false
}
