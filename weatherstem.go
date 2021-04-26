// My local area WeatherSTEM data.
// Sign up for an API key and create a little JSON config file in it.

// main package for weatherstem CLI tool as a command line application
package main

import (
	"crypto/tls"
	"flag"
	"html"
	"log"
	"strconv"

	json "github.com/json-iterator/go"

	"github.com/loraxipam/compassrose"
	haversine "github.com/loraxipam/havers2"

	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	// Version of the configuration file layout
	configSettingsVersion = "3.0"
)

// WeatherInfo struct
// This is the primary structure for weather data which includes the recording station
// info as well as the data series
type WeatherInfo struct {
	WeatherRecord  RecordInfo  `json:"record"`
	WeatherStation StationInfo `json:"station"`
}

// RecordInfo struct
// Currently (June 2020), weatherSTEM has a formatting problem on the output of the JSON
// when a station is "down" -- all numeric scalars become numbers instead of the usual
// string. This kills the unmarshalling so expect errors once in a while.
type RecordInfo struct {
	RecordReadings    []ReadingInfo `json:"readings"`
	LastRainTime      string        `json:"last_rain_time"`
	ReadingsTimestamp string        `json:"time"`
	RecordID          string        `json:"id"`
	RecordHiLo        HiloInfo      `json:"hilo"`
	RecordTimestamp   string        `json:"now"`
	RecordDataDerived uint8         `json:"derived"`
	StationDown       string        `json:"down_since,omitempty"`
}

// StationInfo struct
// This is the basic info about the site which recorded the weather data
type StationInfo struct {
	Domain         DomainInfo   `json:"domain"`
	Cameras        []CameraInfo `json:"cameras"`
	Name           string       `json:"name"`
	Handle         string       `json:"handle"`
	Longitude      string       `json:"lon"`
	Latitude       string       `json:"lat"`
	FacebookID     string       `json:"facebook"`
	TwitterID      string       `json:"twitter"`
	WundergroundID string       `json:"wunderground"`
}

// WeatherData scalars from API's strings are layed out thusly:
// "sensor_type": "Thermometer",
// "sensor_type": "Dewpoint",
// "sensor_type": "Wet Bulb Globe Temperature",
// "sensor_type": "Wind Chill",
// "sensor_type": "Heat Index"
// ---------------------------
// "sensor_type": "Hygrometer"
// ---------------------------
// "sensor_type": "Anemometer",
// "sensor_type": "10 Minute Wind Gust",
// "sensor_type": "Wind Vane",
// ---------------------------
// "sensor_type": "Barometer"
// "sensor_type": "Barometer Tendency",
// ---------------------------
// "sensor_type": "Rain Gauge",
// "sensor_type": "Rain Rate",
// ---------------------------
// "sensor_type": "Solar Radiation Sensor",
// "sensor_type": "UV Radiation Sensor"
type WeatherData struct {
	Label         string          `json:"label"`
	Station       [3]string       `json:"stations"`
	StationTopo   haversine.Coord `json:"topo"`
	StationDist   float64         `json:"distance"`
	Temperature   [5]float64      `json:"temp"`
	Humidity      float64         `json:"humidity"`
	Windspeed     [3]float64      `json:"windspeed"`
	Wind          [2]string       `json:"wind"`
	Pressure      float64         `json:"pressure"`
	PressureTrend string          `json:"ptrend"`
	Rain          [2]float64      `json:"rain"`
	Sun           [2]float64      `json:"sun"`
}

// WeatherUnits are the corresponding measurement units for WeatherData values
type WeatherUnits struct {
	Label       string    `json:"label"`
	Station     [3]string `json:"stations"`
	StationTopo struct {
		Lat string `json:"Lat"`
		Lon string `json:"Lon"`
	} `json:"topo"`
	StationDist   string    `json:"distance"`
	Temperature   [5]string `json:"temp"`
	Humidity      string    `json:"humidity"`
	Windspeed     [3]string `json:"windspeed"`
	Wind          [2]string `json:"wind"`
	Pressure      string    `json:"pressure"`
	PressureTrend string    `json:"ptrend"`
	Rain          [2]string `json:"rain"`
	Sun           [2]string `json:"sun"`
}

// ReadingInfo struct describes each measurement
type ReadingInfo struct {
	ID            string `json:"id"`
	Sensor        string `json:"sensor"`
	SensorType    string `json:"sensor_type"`
	TransmitterID string `json:"transmitter"`
	Unit          string `json:"unit"`
	UnitSymbol    string `json:"unit_symbol"`
	Value         string `json:"value"`
}

// HiloInfo This is at least what comes back with Temp info
// and describes the station's maximum/minimum readings over the latest
// time window, usually 24 hours
type HiloInfo struct {
	Name             string `json:"name"`
	Minimum          string `json:"min"`
	Maximum          string `json:"max"`
	MinimumTimestamp string `json:"min_time"`
	Symbol           string `json:"symbol"`
	MaximumTime      string `json:"max_time"`
	Property         string `json:"property"`
	Type             string `json:"type"`
	Unit             string `json:"unit"`
}

// DomainInfo struct is basically the alias for the individual WeatherSTEM stations
type DomainInfo struct {
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

// CameraInfo struct describes pointers to recent images from the station camera
type CameraInfo struct {
	ImageURL string `json:"image"`
	Name     string `json:"name"`
}

// weatherSTEM API user config settings, ala:
// {"version": "2.0",
// "api_url": "https://volusia.weatherstem.com/api",
// "api_key": "happy3solar9fly",
// "stations": ["ponceinlet","fswndaytonabch"]
// "me": {"lat":29.13,"lon":-80.95}
// }
// See weatherstem API page for details.
// This is version 2. -- Added "Me"
type configSettings struct {
	Version  string          `json:"version"`
	URL      string          `json:"api_url"`
	Key      string          `json:"api_key"`
	Stations []string        `json:"stations"`
	Me       haversine.Coord `json:"me,omitempty"`
}

// PopulateWeatherData accepts the raw result and it returns the converted structured data
// This is a rather static implementation. Bigger brains do this better.
func PopulateWeatherData(winfo *WeatherInfo, rose bool) (wdata WeatherData, wunits WeatherUnits) {
	wdata.Label = "data"
	wdata.Station[0] = winfo.WeatherStation.Handle
	wdata.Station[1] = winfo.WeatherStation.Name
	wdata.Station[2] = winfo.WeatherRecord.ReadingsTimestamp
	wdata.StationTopo.Lat, _ = strconv.ParseFloat(winfo.WeatherStation.Latitude, 64)
	wdata.StationTopo.Lon, _ = strconv.ParseFloat(winfo.WeatherStation.Longitude, 64)
	wdata.StationTopo.Calc()
	wdata.StationDist = 2.4
	wunits.Label = "units"
	wunits.Station[0] = winfo.WeatherStation.Handle
	wunits.Station[1] = winfo.WeatherStation.Name
	wunits.Station[2] = winfo.WeatherRecord.ReadingsTimestamp
	wunits.StationTopo.Lat = "&deg;"
	wunits.StationTopo.Lon = "&deg;"
	// now loop through the readings and do the conversions
	for _, val := range winfo.WeatherRecord.RecordReadings {
		if val.SensorType == "Thermometer" { // Temps
			wdata.Temperature[0], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Temperature[0] = val.UnitSymbol
		} else if val.SensorType == "Dewpoint" {
			wdata.Temperature[1], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Temperature[1] = val.UnitSymbol
		} else if val.SensorType == "Wet Bulb Globe Temperature" {
			wdata.Temperature[2], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Temperature[2] = val.UnitSymbol
		} else if val.SensorType == "Wind Chill" {
			wdata.Temperature[3], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Temperature[3] = val.UnitSymbol
		} else if val.SensorType == "Heat Index" {
			wdata.Temperature[4], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Temperature[4] = val.UnitSymbol
		} else if val.SensorType == "Hygrometer" { // Humidity
			wdata.Humidity, _ = strconv.ParseFloat(val.Value, 64)
			wunits.Humidity = val.UnitSymbol
		} else if val.SensorType == "Anemometer" { // Wind
			wdata.Windspeed[0], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Windspeed[0] = val.UnitSymbol
		} else if val.SensorType == "10 Minute Wind Gust" {
			wdata.Windspeed[1], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Windspeed[1] = val.UnitSymbol
		} else if val.SensorType == "Wind Vane" {
			wdata.Windspeed[2], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Windspeed[2] = val.UnitSymbol
			if rose {
				wdata.Wind[0], wdata.Wind[1] = compassrose.DegreeToHeading(float32(wdata.Windspeed[2]), 3, true)
			} else {
				wdata.Wind[0], wdata.Wind[1] = compassrose.DegreeToHeading(float32(wdata.Windspeed[2]), 3, false)
			}
		} else if val.SensorType == "Barometer" { // Pressure
			wdata.Pressure, _ = strconv.ParseFloat(val.Value, 64)
			wunits.Pressure = val.UnitSymbol
		} else if val.SensorType == "Barometer Tendency" {
			wdata.PressureTrend = val.Value
			wunits.PressureTrend = val.UnitSymbol
		} else if val.SensorType == "Rain Gauge" { // Rain
			wdata.Rain[0], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Rain[0] = val.UnitSymbol
		} else if val.SensorType == "Rain Rate" {
			wdata.Rain[1], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Rain[1] = val.UnitSymbol
		} else if val.SensorType == "Solar Radiation Sensor" { // Sun
			wdata.Sun[0], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Sun[0] = val.UnitSymbol
		} else if val.SensorType == "UV Radiation Sensor" {
			wdata.Sun[1], _ = strconv.ParseFloat(val.Value, 64)
			wunits.Sun[1] = val.UnitSymbol
		} // else ignore the unknown
	}

	return wdata, wunits
}

// get config settings from the usual suspect files. Look in current directory, HOME or .config
func findConfigSettings(config *configSettings) (err error) {
	var usualFiles [3]string
	if home, exists := os.LookupEnv("HOME"); exists {
		usualFiles[0] = "weatherstem.json"
		usualFiles[1] = home + "/.weatherstem.json"
		usualFiles[2] = home + "/.config/weatherstem.json"
	} else {
		usualFiles[0] = "weatherstem.json"
	}
	for _, c := range usualFiles {
		err = config.getConfigSettings(c)
		if err == nil {
			return err
		}
	}

	return err
}

// get API user config settings from a file
func (config *configSettings) getConfigSettings(inputFile string) (err error) {
	readFile, err := os.Open(inputFile)
	if err != nil {
		// trying to open a non-existent file is not a panic
		return err
	}
	defer readFile.Close()

	configJSON, err := ioutil.ReadAll(readFile)
	if err != nil {
		log.Panicln("Cannot read config", inputFile)
	}

	var configVersion string

	// Confirm config version
	if strings.Contains(string(configJSON), "version") {
		v1 := strings.SplitAfter(string(configJSON), `"version"`)
		v1 = strings.Split(v1[1], `"`)
		configVersion = v1[1]
	} else {
		log.Panicln("No version in config file.", inputFile)
	}

	if configVersion == configSettingsVersion {
		err = json.Unmarshal(configJSON, &config)
		if err != nil {
			log.Panicln("Cannot unmarshal config", inputFile)
		}
	} else if configVersion <= configSettingsVersion {
		log.Printf("WARNING: Using a version %s config file in a version %s app.\n", configVersion, configSettingsVersion)
		log.Printf("Version 2 added your geolocation. Your location could become NYC.\n")
		log.Printf("Version 3 uses the Aug 2020 API v1 'station@domain.weatherstem.com' syntax.\n")
		err = json.Unmarshal(configJSON, &config)
		if err != nil {
			log.Panicln("Cannot unmarshal config", inputFile)
		}
		if config.Me.Lat == 0.0 {
			config.Me.Lat = 40.7678
		}
		if config.Me.Lon == 0.0 {
			config.Me.Lon = -73.9814
		}
	} else {
		log.Panicf("Config version mismatch, %v should be %v\n", configVersion, configSettingsVersion)
	}

	config.Me.Calc()

	return err
}

// get weather data from some file, if you want to test things locally
func getWeatherInfoFromSomeFile(inputFile string) ([]byte, error) {
	readFile, _ := os.Open(inputFile)
	defer readFile.Close()
	usualCallBody, err := ioutil.ReadAll(readFile)
	return usualCallBody, err
}

// get weather data from the web site
func getWeatherInfoFromWeb(c *configSettings) ([]byte, error) {

	// We need a TLS session
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// We need a client for the TLS session
	client := &http.Client{Transport: transport}

	// We need a request URL which we get from our config file's api_url
	// something like 'https://volusia.weatherstem.com/api'
	apiURL := c.URL
	// and the contents of the request. My local station data from the config file's stations array. Je suis hackeur.
	requestBody := `{"api_key":"` + c.Key + `","stations":["` + strings.Join(c.Stations, `","`) + `"]}`
	// requestBody is sorta like: {"api_key":"polyshazbotmicrofish","stations":["ponceinlet","fswndaytonabch"]}

	body := strings.NewReader(requestBody)

	// Make the call
	responseBody, err := client.Post(apiURL, "application/json", body)
	if err != nil {
		return nil, err
	}

	// Close the session once we're done
	defer responseBody.Body.Close()

	// Now parse the result
	apiResponse, err := ioutil.ReadAll(responseBody.Body)

	return apiResponse, err
}

// PrintWeatherDataJSON shows the data and the measurement units for a station
func (data *WeatherData) PrintWeatherDataJSON(units *WeatherUnits) {
	var jdata, junits []byte
	var err error
	jdata, err = json.Marshal(data)
	if err != nil {
		log.Println("Cannot marshal weather info", err)
		// return err
	}

	junits, err = json.Marshal(units)
	if err != nil {
		log.Println("Cannot marshal unit info", err)
		// return err
	}

	fmt.Printf("%s\n", string(jdata))
	fmt.Printf("%s\n", string(junits))
}

// PrintWeatherInfoJSON shows the data only for a station. No units.
func (data *WeatherInfo) PrintWeatherInfoJSON() {
	var jdata []byte
	var err error
	jdata, err = json.Marshal(data)
	if err != nil {
		log.Println("Cannot marshal weather data", err)
		// return err
	}

	fmt.Printf("%s\n", string(jdata))
}

// WBGTFlag returns the "danger" flag for a given wet bulb globe temperature
func WBGTFlag ( temp float64 ) (flag string) {
	var runeflag = []rune("⚊⚌☰⚑")
	switch {
	case (temp < 82.0):
		return " "
	case (temp < 87.0):
		return string(runeflag[0])
	case (temp < 90.0):
		return string(runeflag[1])
	case (temp < 92.0):
		return string(runeflag[2])
	default:
		return string(runeflag[3])
	}

}

// PrintWeatherData shows the (REAL basic) data for a station
func (data *WeatherData) PrintWeatherData() {

	fmt.Println(data.Station[1], "("+data.Station[0]+")", data.Station[2], data.StationDist)
	fmt.Println(" ", " T:", data.Temperature[0], "DP:", data.Temperature[1], "H:", data.Humidity)
	fmt.Println(WBGTFlag(data.Temperature[2]), "WB:", data.Temperature[2], "WC:", data.Temperature[3], "HI:", data.Temperature[4])
	fmt.Println(" ", " P:", data.Pressure, data.PressureTrend)
	fmt.Println(" ", " W:", data.Windspeed[0], data.Windspeed[1], "gust", "("+strconv.FormatFloat(data.Windspeed[2], 'f', 0, 64)+"°", data.Wind[1]+")")
	fmt.Println(" ", " R:", data.Rain[0], "gauge", data.Rain[1], "rate")
}

// PrintWeatherDataUnits shows the data for a station along with its units
func (data *WeatherData) PrintWeatherDataUnits(wu *WeatherUnits) {

	// Many of the unit strings are HTML-escaped
	fmt.Printf("%s (%s) %.2f%s %s\n", data.Station[1], data.Station[0], data.StationDist, wu.StationDist, data.Station[2])
	fmt.Printf(" T: %-.1f%s DP: %-.1f%s H: %.1f%s\n", data.Temperature[0], html.UnescapeString(wu.Temperature[0]), data.Temperature[1], html.UnescapeString(wu.Temperature[1]), data.Humidity, "%")
	fmt.Printf("WB: %-.1f%s %s WC: %-.1f%s HI: %-.1f%s\n", data.Temperature[2], html.UnescapeString(wu.Temperature[2]), WBGTFlag(data.Temperature[2]),data.Temperature[3], html.UnescapeString(wu.Temperature[3]), data.Temperature[4], html.UnescapeString(wu.Temperature[4]))
	fmt.Printf(" P: %.3f%s [%.2fmbar] %v\n", data.Pressure, wu.Pressure, data.Pressure*33.86386, data.PressureTrend) // Major assumption here!
	fmt.Printf(" W: %.1f%s %.1f%s gust, %v%v %s\n", data.Windspeed[0], wu.Windspeed[0], data.Windspeed[1], html.UnescapeString(wu.Windspeed[1]), data.Windspeed[2], html.UnescapeString(wu.Windspeed[2]), data.Wind[1])
	fmt.Printf(" R: %.2f%s %.2f%s\n", data.Rain[0], wu.Rain[0], data.Rain[1], wu.Rain[1])
}

// main body function
func main() {

	var (
		weatherBytes                             []byte			// The API returns a JSON array of stations with their data
		err                                      error
		weatherArr                               []WeatherInfo		// The structured API data
		myConfig                                 configSettings		// Your API user info, location and local WeatherSTEM sites
		outputJSON, outputOrig, rose, kilo, mile, lite bool		// Some command line flags
	)

	// Get the commandline flags
	flag.BoolVar(&outputJSON, "json", false, "Output cooked data as JSON")
	flag.BoolVar(&kilo, "kilo", false, "Output station distances in kilometers")
	flag.BoolVar(&mile, "mile", false, "Output station distances in statute miles")
	flag.BoolVar(&lite, "lite", false, "Output lightweight cooked data")
	flag.BoolVar(&outputOrig, "orig", false, "Output original API results")
	flag.BoolVar(&rose, "rose", false, "Output boring compass rose directions")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Println("Current WBGT flags:")
		fmt.Println("   <82°F       - normal")
		fmt.Println(" ⚊ 82°F - 87°F - Level 1")
		fmt.Println(" ⚌ 87°F - 90°F - Level 2")
		fmt.Println(" ☰ 90°F - 92°F - Level 3")
		fmt.Println(" ⚑ >92°F       - Level 4")
		os.Exit(0)
	}

	// Get API and stations from the configuration file in the current directory or HOME directory
	err = findConfigSettings(&myConfig)
	if err != nil {
		log.Println("Config file not found. It should look like this and be in 'weatherstem.json', either in the current or in your $HOME/.config directory.")
		log.Println(`{"version":"3.0","api_url":"https://api.weatherstem.com/api","api_key":"yourApiKey","stations":["station1@domain.weatherstem.com","stationX@domain.weatherstem.com"],"me":{"lat":43.14,"lon":-111.275}}`)
		os.Exit(3)
	}

	// Get local WeatherSTEM data
	weatherBytes, err = getWeatherInfoFromWeb(&myConfig)
	if err != nil {
		log.Println("Call to API failed.", err)
		os.Exit(1)
	}

	// Parse returned data into basic structs
	err = json.Unmarshal(weatherBytes, &weatherArr)
	if err != nil {
		log.Println("Cannot unmarshal API results.")
		log.Println(string(weatherBytes))
		os.Exit(2)
	}

	// Convert stringy structs into scalars
	dataArr := make([]WeatherData, len(weatherArr))
	unitArr := make([]WeatherUnits, len(weatherArr))
	for idx, stationData := range weatherArr {
		dataArr[idx], unitArr[idx] = PopulateWeatherData(&stationData, rose)
		if kilo {
			dataArr[idx].StationDist = haversine.DistanceKm(myConfig.Me, dataArr[idx].StationTopo)
			unitArr[idx].StationDist = "km"
		} else if mile {
			dataArr[idx].StationDist = haversine.DistanceMi(myConfig.Me, dataArr[idx].StationTopo)
			unitArr[idx].StationDist = "mi"
		} else {
			dataArr[idx].StationDist = haversine.DistanceNM(myConfig.Me, dataArr[idx].StationTopo)
			unitArr[idx].StationDist = "NM"
		}
	}

	// Show the original raw info
	if outputOrig {
		for _, origInfo := range weatherArr {
			origInfo.PrintWeatherInfoJSON()
		}
	} else {

		// Show the cooked data
		for i := 0; i < len(dataArr); i++ {
			if outputJSON {
				dataArr[i].PrintWeatherDataJSON(&unitArr[i])
			} else {
				if lite {
					dataArr[i].PrintWeatherData()
				} else {
					dataArr[i].PrintWeatherDataUnits(&unitArr[i])
				}
			}
		}
	}

	// Add your other fun stuff here.
}
