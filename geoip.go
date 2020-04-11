package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// This is the format returned by: https://ipinfo.io/w.x.y.z/json
type ipInfoResult struct {
	IP       string
	Hostname string
	City     string
	Region   string
	Country  string
	Loc      string
	Postal   string
	Org      string
	Distance float64
	ErrMsg   error
}

/*
getIpInfo issues a web query to ipinfo.io
The JSON result is converted to an ipInfoResult struct
Args:
	ip: an IPv4 address

Returns:
	an ipInfoResult struct containing the information returned by the service
*/
func getIPInfo(ip string) (*ipInfoResult, error) {
	var obj ipInfoResult

	api := "/json"
	if 0 == len(ip) {
		api = "json"
	}
	url := "https://ipinfo.io/" + ip + api
	resp, err := http.Get(url)
	if err != nil {
		return &obj, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &obj, err
	}

	if strings.Contains(string(body), "Rate limit exceeded") {
		err := fmt.Errorf("Error for '%s', %s", url, string(body))
		return new(ipInfoResult), err
	}

	json.Unmarshal(body, &obj)
	return &obj, nil
}

func validateLocation(src net.Conn, geoIP ipInfoResult) string {
	slots := strings.Split(src.RemoteAddr().String(), ":")
	remoteIP := slots[0]
	remoteIPInfo, _ := getIPInfo(remoteIP)

	var miles float64
	var lat1, lon1, lat2, lon2 float64
	var err error
	if len(geoIP.Loc) > 0 {
		lat1, lon1, err = latlon2coord(remoteIPInfo.Loc)
		if err != nil {
			return "remote-LatLon"
		}
		lat2, lon2, err = latlon2coord(geoIP.Loc)
		if err != nil {
			return "geoIP-LatLon"
		}

		miles = HaversineDistance(lat1, lon1, lat2, lon2)
		fmt.Println("Required Distance:", geoIP.Distance)
		fmt.Println("  Actual Distance:", miles)
	}

	distanceStr := fmt.Sprintf("%.2f", miles)
	fmt.Println("distance: ", distanceStr)
	fmt.Println(remoteIPInfo)

	mismatch := ""
	if len(geoIP.City) > 0 && geoIP.City != strings.ToUpper(remoteIPInfo.City) {
		mismatch = "City"
	}
	if len(geoIP.Region) > 0 && geoIP.Region != strings.ToUpper(remoteIPInfo.Region) {
		mismatch = "Region"
	}
	if len(geoIP.Country) > 0 && geoIP.Country != strings.ToUpper(remoteIPInfo.Country) {
		mismatch = "Country"
	}
	if len(geoIP.Loc) > 0 && miles > geoIP.Distance {
		mismatch = "Distance"
	}
	return mismatch
}

/*
latlon2coord converts a string such as "36.0525,-79.107" to a tuple of floats

Args:
	latlon: a string in "lat, lon" format

Returns:
	a tuple in (float64, float64) format
*/
func latlon2coord(latlon string) (float64, float64, error) {
	slots := strings.Split(latlon, ",")
	lat, err := strconv.ParseFloat(slots[0], 64)
	if err != nil {
		err = fmt.Errorf("Error converting latitude to float for: %s", latlon)
	}
	lon, err := strconv.ParseFloat(slots[1], 64)
	if err != nil {
		err = fmt.Errorf("Error converting longitude to float for: %s", latlon)
	}
	return lat, lon, err
}

// adapted from: https://gist.github.com/cdipaolo/d3f8db3848278b49db68
// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// HaversineDistance returns the distance (in miles) between two points of
//	 a given longitude and latitude relatively accurately (using a spherical
//	 approximation of the Earth) through the Haversin Distance Formula for
//	 great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// http://en.wikipedia.org/wiki/Haversine_formula
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64

	piRad := math.Pi / 180
	la1 = lat1 * piRad
	lo1 = lon1 * piRad
	la2 = lat2 * piRad
	lo2 = lon2 * piRad

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	meters := 2 * r * math.Asin(math.Sqrt(h))
	miles := meters / 1609.344
	return miles
}
