package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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
func getIPInfo(ip string) (ipInfoResult, error) {
	var obj ipInfoResult

	api := "/json"
	if 0 == len(ip) {
		api = "json"
	}
	url := "https://ipinfo.io/" + ip + api
	resp, err := http.Get(url)
	if err != nil {
		return obj, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return obj, err
	}

	if strings.Contains(string(body), "Rate limit exceeded") {
		err := fmt.Errorf("Error for '%s', %s", url, string(body))
		empty := ipInfoResult{}
		return empty, err
	}

	json.Unmarshal(body, &obj)
	return obj, nil
}

func validateLocation(localGeoIP ipInfoResult, remoteGeoIP ipInfoResult, restrictionsGeoIP ipInfoResult) (string, string) {
	var distanceCalc string

	if 0 == len(localGeoIP.Loc) {
		return fmt.Sprintf("localGeoIP '%s' does not have lat,lon", localGeoIP.IP), ""
	}
	if 0 == len(remoteGeoIP.Loc) {
		return fmt.Sprintf("remoteGeoIP '%s' does not have lat,lon", remoteGeoIP.IP), ""
	}

	var miles float64
	var lat1, lon1, lat2, lon2 float64
	var err error

	if restrictionsGeoIP.Distance > 0 {
		lat1, lon1, err = latlon2coord(remoteGeoIP.Loc)
		if err != nil {
			return "", "remote-LatLon coordinates"
		}

		if len(restrictionsGeoIP.Loc) > 0 { // location and distance
			lat2, lon2, err = latlon2coord(restrictionsGeoIP.Loc)
			if err != nil {
				return "restrictions.GeoIP-LatLon coordinates", ""
			}
		} else { // distance only
			lat2, lon2, err = latlon2coord(localGeoIP.Loc)
			if err != nil {
				return "localGeoIP-LatLon coordinates", ""
			}
		}
		miles = HaversineDistance(lat1, lon1, lat2, lon2)
		milesDiff := math.Abs(miles - restrictionsGeoIP.Distance)
		distanceCalc = fmt.Sprintf("Current Distance: %.2f; Maximum Distance: %.2f; Difference: %.2f", miles, restrictionsGeoIP.Distance, milesDiff)
	}

	mismatch := ""
	if len(restrictionsGeoIP.City) > 0 && strings.ToUpper(restrictionsGeoIP.City) != strings.ToUpper(remoteGeoIP.City) {
		mismatch = fmt.Sprintf("[%s] Forbidden City: %s", remoteGeoIP.IP, remoteGeoIP.City)
	} else if len(restrictionsGeoIP.Region) > 0 && strings.ToUpper(restrictionsGeoIP.Region) != strings.ToUpper(remoteGeoIP.Region) {
		mismatch = fmt.Sprintf("[%s] Forbidden Region: %s", remoteGeoIP.IP, remoteGeoIP.Region)
	} else if len(restrictionsGeoIP.Country) > 0 && strings.ToUpper(restrictionsGeoIP.Country) != strings.ToUpper(remoteGeoIP.Country) {
		mismatch = fmt.Sprintf("[%s] Forbidden Country: %s", remoteGeoIP.IP, remoteGeoIP.Country)
	} else if restrictionsGeoIP.Distance > 0 && miles > restrictionsGeoIP.Distance {
		// mismatch = fmt.Sprintf("[%s] Forbidden Distance: %.2f", remoteGeoIP.IP, miles)
		mismatch = fmt.Sprintf("[%s] DENY;", remoteGeoIP.IP)
	}
	return mismatch, distanceCalc
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
