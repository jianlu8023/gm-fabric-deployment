package geoip

// CountryInfo 存储国家信息的结构体
type CountryInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CityInfo 存储城市信息的结构体
type CityInfo struct {
	Name string `json:"name"`
}

// LocationInfo 存储位置信息的结构体
type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	TimeZone  string  `json:"time_zone"`
}

// IPInfo 存储完整的IP信息结构体
type IPInfo struct {
	IP        string       `json:"ip"`
	Country   CountryInfo  `json:"country"`
	City      CityInfo     `json:"city"`
	Location  LocationInfo `json:"location"`
	Postal    string       `json:"postal"`
	Continent string       `json:"continent"`
}
