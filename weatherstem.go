// My local area WeatherSTEM data.
// Sign up for an API key and create a little JSON config file in it.

package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"html"
	"log"
	"strconv"

	"github.com/loraxipam/haversine"

	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/loraxipam/compassrose"
)

const (
	configSettingsVersion = "1.0"
)

// WeatherInfo struct
type WeatherInfo struct {
	WeatherRecord  RecordInfo  `json:"record"`
	WeatherStation StationInfo `json:"station"`
}

// RecordInfo struct
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
	Station       [3]string
	StationTopo   haversine.Coord
	Temperature   [5]float64
	Humidity      float64
	Windspeed     [3]float64
	Wind          [2]string
	Pressure      float64
	PressureTrend string
	Rain          [2]float64
	Sun           [2]float64
}

// WeatherUnits are the measurement units for the WeatherData values
type WeatherUnits struct {
	Station       [3]string
	StationTopo   string
	Temperature   [5]string
	Humidity      string
	Windspeed     [3]string
	Wind          [2]string
	Pressure      string
	PressureTrend string
	Rain          [2]string
	Sun           [2]string
}

// ReadingInfo struct
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

// DomainInfo struct
type DomainInfo struct {
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

// CameraInfo struct
type CameraInfo struct {
	ImageURL string `json:"image"`
	Name     string `json:"name"`
}

// weatherSTEM API user config settings, ala:
// {"version": "1.0",
// "api_url": "https://volusia.weatherstem.com/api",
// "api_key": "happy3solar9fly",
// "stations": ["ponceinlet","fswndaytonabch"]
// }
// See weatherstem API page for details.
type configSettings struct {
	Version  string   `json:"version"`
	URL      string   `json:"api_url"`
	Key      string   `json:"api_key"`
	Stations []string `json:"stations"`
}

// PopulateWeatherData accepts the raw result and it returns the converted structured data
// This is a rather static implementation. Bigger brains do this better.
func PopulateWeatherData(winfo *WeatherInfo) (wdata WeatherData, wunits WeatherUnits) {
	wdata.Station[0] = winfo.WeatherStation.Handle
	wdata.Station[1] = winfo.WeatherStation.Name
	wdata.Station[2] = winfo.WeatherRecord.ReadingsTimestamp
	wdata.StationTopo.Lat, _ = strconv.ParseFloat(winfo.WeatherStation.Latitude, 64)
	wdata.StationTopo.Lon, _ = strconv.ParseFloat(winfo.WeatherStation.Longitude, 64)
	wunits.StationTopo = "°"
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
			wdata.Wind[0], wdata.Wind[1] = compassrose.DegreeToHeading(float32(wdata.Windspeed[2]), 3, false)
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
		}

	}

	return wdata, wunits
}

// get config settings from the usual suspect files
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
	defer readFile.Close()
	if err != nil {
		// trying to open a non-existent file is not a panic
		return err
	}

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
	} else {
		log.Panicf("Config version mismatch, %v should be %v\n", configVersion, configSettingsVersion)
	}

	return err
}

// get weather data from some file, if you want to test things locally
func getWeatherInfoFromSomeFile(inputFile string) ([]byte, error) {
	readFile, _ := os.Open(inputFile)
	usualCallBody, err := ioutil.ReadAll(readFile)
	defer readFile.Close()
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

	// requestBody = `{"api_key":"polyshazbotmicrofish","stations":["ponceinlet","fswndaytonabch"]}`
	body := strings.NewReader(requestBody)

	// Make the call
	responseBody, error := client.Post(apiURL, "application/json", body)

	if error != nil {
		return nil, error
	}
	// Close the session once we're done
	defer responseBody.Body.Close()

	// Now parse the result
	apiResponse, error := ioutil.ReadAll(responseBody.Body)

	return apiResponse, error
}

// PrintWeatherDataJSON shows the data for a station
func (data *WeatherData) PrintWeatherDataJSON() {
	var jdata []byte
	jdata, err := json.Marshal(data)
	if err != nil {
		log.Println("Cannot marshal data")
	}

	fmt.Printf("%s\n", string(jdata))
}

// PrintWeatherInfoJSON shows the data for a station
func (data *WeatherInfo) PrintWeatherInfoJSON() {
	var jdata []byte
	jdata, err := json.Marshal(data)
	if err != nil {
		log.Println("Cannot marshal data")
	}

	fmt.Printf("%s\n", string(jdata))
}

// PrintWeatherData shows the data for a station
func (data *WeatherData) PrintWeatherData() {

	fmt.Println(data.Station[1], "("+data.Station[0]+")", data.Station[2])
	fmt.Println(" ", " T:", data.Temperature[0], "DP:", data.Temperature[1], "H:", data.Humidity)
	fmt.Println(" ", "WB:", data.Temperature[2], "WC:", data.Temperature[3], "HI:", data.Temperature[4])
	fmt.Println(" ", " P:", data.Pressure, data.PressureTrend)
	fmt.Println(" ", " W:", data.Windspeed[0], data.Wind[1], "("+strconv.FormatFloat(data.Windspeed[2], 'f', 0, 64)+"°)", data.Windspeed[1], "gust")
	fmt.Println(" ", " R:", data.Rain[0], "gauge", data.Rain[1], "rate")
}

// PrintWeatherDataUnits shows the data for a station along with its units
func (data *WeatherData) PrintWeatherDataUnits(wu *WeatherUnits) {

	fmt.Println(data.Station[1], "("+data.Station[0]+")", data.Station[2])
	fmt.Printf(" T: %-.1f%s DP: %-.1f%s H: %.1f%s\n", data.Temperature[0], html.UnescapeString(wu.Temperature[0]), data.Temperature[1], html.UnescapeString(wu.Temperature[1]), data.Humidity, "%")
	fmt.Printf("WB: %-.1f%s WC: %-.1f%s HI: %-.1f%s\n", data.Temperature[2], html.UnescapeString(wu.Temperature[2]), data.Temperature[3], html.UnescapeString(wu.Temperature[3]), data.Temperature[4], html.UnescapeString(wu.Temperature[4]))
	fmt.Printf(" P: %.3f%s [%.2fmbar] %v\n", data.Pressure, wu.Pressure, data.Pressure*33.86386, data.PressureTrend) // Major assumption here!
	fmt.Printf(" W: %.1f%s %s %v%v %.1f%s gust\n", data.Windspeed[0], wu.Windspeed[0], data.Wind[1], data.Windspeed[2], html.UnescapeString(wu.Windspeed[2]), data.Windspeed[1], html.UnescapeString(wu.Windspeed[1]))
	fmt.Printf(" R: %.2f%s %.2f%s\n", data.Rain[0], wu.Rain[0], data.Rain[1], wu.Rain[1])
}

func main() {

	var (
		weatherBytes           []byte
		err                    error
		weatherArr             []WeatherInfo
		myConfig               configSettings
		outputJSON, outputOrig bool
	)

	// Get the commandline flags
	flag.BoolVar(&outputJSON, "json", false, "Output cooked data as JSON")
	flag.BoolVar(&outputOrig, "orig", false, "Output original API results")
	flag.Parse()

	// Get API and stations from the configuration file in the current directory or HOME directory
	err = findConfigSettings(&myConfig)
	if err != nil {
		log.Println("Config file not found. It should look like this and be in 'weatherstem.json', either in the current or in your $HOME/.config directory.")
		log.Println(`{"version":"1.0","api_url":"https://domain.weatherstem.com/api","api_key":"yourApiKey","stations":["station1","stationX"]}`)
		os.Exit(3)
	}

	// Get local WeatherSTEM data
	// weatherBytes, _ := getPonceInfoFromFile()
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
		dataArr[idx], unitArr[idx] = PopulateWeatherData(&stationData)
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
				dataArr[i].PrintWeatherDataJSON()
			} else {
				dataArr[i].PrintWeatherDataUnits(&unitArr[i])
			}
		}
	}

	// Add your other fun stuff here.
}
