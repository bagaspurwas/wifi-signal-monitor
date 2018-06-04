package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

const (
	//ConfigFile name without extension
	ConfigFile = "config"
	//ConfigPath absolute
	ConfigPath = "/etc/wifimonitor/"
)

// WirelessNetwork defines a Wireless Network with
// SSOD, AP Mac Address, operation frequency, and signal strength
type WirelessNetwork struct {
	SSID   string
	MAC    string
	freq   string
	signal float32
}

type influxdbConf struct {
	host            string
	port            string
	username        string
	password        string
	database        string
	measurement     string
	retentionPolicy string
}

// Configuration file
type Configuration struct {
	uniqueID      string
	wlanInterface string
	location      string
	threads       int
	influxdbConf
}

func handleError(err error) {
	if err != nil {
		log.Println("Error!")
	}
}

// ScanWiFi function scan wireless network using iw
// and return list of WirelessNetwork
func ScanWiFi(wlanInterface string) []WirelessNetwork {
	WiFiList := make([]WirelessNetwork, 0, 100)

	log.Println("Scanning Wireless Network on " + wlanInterface)

	args := "iw " + wlanInterface + " scan | grep -e 'SSID:\\|signal\\|freq:\\|BSS [a-f0-9]'"
	cmd := exec.Command("/bin/bash", "-c", args)
	stdout, err := cmd.StdoutPipe()

	handleError(err)

	cmd.Start()

	var line = 1
	var signalStr float64
	var freq string
	var MAC string

	buffer := bufio.NewScanner(stdout)
	buffer.Split(bufio.ScanLines)

	for buffer.Scan() {
		bufField := strings.Fields(buffer.Text())
		//Iterate through the line and add corresponding line
		// to WiFiList param

		switch line % 4 {
		case 0:
			SSID := ""
			if len(bufField) == 1 {
				SSID = ""
			} else {
				for _, v := range bufField[1:] {
					SSID += v
				}
			}
			WiFiList = append(WiFiList, WirelessNetwork{SSID: SSID,
				MAC: MAC, signal: float32(signalStr), freq: freq})
		case 1:
			MAC = bufField[1][:17]
		case 2:
			freq = bufField[1]

		case 3:
			signalStr, err = strconv.ParseFloat(bufField[1], 32)
			handleError(err)
		}

		line++

	}

	return WiFiList
}

func loadConfig() Configuration {
	viper.SetConfigName(ConfigFile)
	viper.AddConfigPath(ConfigPath)
	err := viper.ReadInConfig()
	//add panic later
	handleError(err)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed: ", e.Name)
	})
	viper.SetConfigType("yaml")

	thisHost := ""
	wlanInterface := ""
	//If uniqueID is not defined, use hostname instead
	if viper.GetString("probeNode.uniqueID") == "" {
		thisHost, err = os.Hostname()
		handleError(err)
	} else {
		thisHost = viper.GetString("probeNode.uniqueID")
	}

	if viper.GetString("probeNode.wlanInterface") == "" {
		wlanName, err := exec.Command("/bin/bash", "-c",
			"cat /proc/net/wireless | sed -n '3p' | grep -Eo '^[a-z0-9 ]+' | tr -d '\n'").Output()
		handleError(err)
		wlanInterface = string(wlanName)[:]
	} else {
		wlanInterface = viper.GetString("probeNode.wlanInterface")
	}

	return Configuration{
		uniqueID:      thisHost,
		wlanInterface: wlanInterface,
		location:      viper.GetString("probNode.location"),
		threads:       viper.GetInt("probeNode.threaads"),
		influxdbConf: influxdbConf{
			viper.GetString("influxdb.host"),
			viper.GetString("influxdb.port"),
			viper.GetString("influxdb.username"),
			viper.GetString("influxdb.password"),
			viper.GetString("influxdb.database"),
			viper.GetString("influxdb.measurement"),
			viper.GetString("influxdb.retentionPolicy"),
		},
	}
}

func writeInfluxDB(clnt client.Client, config Configuration, WiFi WirelessNetwork) {
	//Load config
	dbname := config.influxdbConf.database
	retentionPolicy := config.influxdbConf.retentionPolicy
	measurement := config.influxdbConf.measurement
	uniqueID := config.uniqueID
	location := config.location
	// write data to database
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        dbname,
		Precision:       "s",
		RetentionPolicy: retentionPolicy,
	})

	handleError(err)

	data := map[string]interface{}{
		"Signal": WiFi.signal,
	}

	tags := map[string]string{
		"Hostname":    uniqueID,
		"SSID":        WiFi.SSID,
		"MAC Address": WiFi.MAC,
		"Frequency":   WiFi.freq,
		"Location":    location,
	}

	pt, err := client.NewPoint(measurement, tags, data, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	bp.AddPoint(pt)

	if err := clnt.Write(bp); err != nil {
		log.Fatal(err)
	}
}

func main() {
	//Catch signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGQUIT)
	go func() {
		s := <-sig
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
	/*Start Program
	 */
	//Load config file
	config := loadConfig()
	dbaddr := config.influxdbConf.host + ":" + config.influxdbConf.port
	dbusername := config.influxdbConf.username
	dbpassword := config.influxdbConf.password
	wlanInterface := config.wlanInterface
	//threads := config.threads
	log.Println(time.Now())
	log.Println("Starting InfluxDB Client...")
	log.Println("InfluxDB Server at " + dbaddr)
	//Create a new HTTP client sending request to InfluxDB
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     dbaddr,
		Username: dbusername,
		Password: dbpassword,
	})

	handleError(err)

	for {
		WiFiList := ScanWiFi(wlanInterface)
		//Add
		for _, WiFi := range WiFiList {
			log.Println(WiFi)
			writeInfluxDB(clnt, config, WiFi)
		}
	}
}
