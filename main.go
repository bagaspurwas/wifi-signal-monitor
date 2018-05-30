package main

import (
	"os"
	"bufio"
	"strings"
	"os/exec"
	"strconv"
	"log"
	"time"
	"os/signal"
	"syscall"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
	"github.com/fsnotify/fsnotify"
)


const (
	CONFIG_FILE = "config"
	CONFIG_PATH = "/home/bapung/Documents/+Project/WiFiSSIDMonitor"
)

//Wifi Information Struct
type WirelessNetwork struct {
	SSID	string		`json:"ssid"`
	MAC		string	
	freq	string		`json:"freq"`
	signal	float32		`json:"signal"`
	country	string		`json:"country"`
}


type conf_influxdb struct {
	host			string
	port			string
	username		string
	password		string
	database		string	
	measurement		string
	retentionPolicy string
}


type Configuration struct {
	probe_hostname string
	conf_influxdb
}


func handleError(err error) {
	if err != nil {
		
	}
}


func ScanWiFi() []WirelessNetwork {
	WiFiList := make([]WirelessNetwork,0,100);
	log.Println("Scanning Wireless Network")
	//args := []string("scan", "wlp3s0")
	cmd := exec.Command("/bin/bash", "-c", 
		"iw wlp3s0 scan | grep -e 'SSID:\\|signal\\|freq:\\|BSS [a-f0-9]'") 
	stdout, err := cmd.StdoutPipe()
	handleError(err)

	cmd.Start()

	var line = 1
	var signalStr float64 = 0.0
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
				MAC: MAC, signal: float32(signalStr) ,freq:freq})
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
	viper.SetConfigName(CONFIG_FILE)
	viper.AddConfigPath(CONFIG_PATH)
	err := viper.ReadInConfig()
	//add panic later
	handleError(err)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed: ", e.Name)
	})
	viper.SetConfigType("yaml")
	
	thisHost := ""

	if (viper.GetString("probeNode.hostname") == "") {
		thisHost, err = os.Hostname()
		handleError(err)
	} else {
		thisHost = viper.GetString("probeNode.hostname") 
	}

	return Configuration{
		probe_hostname : thisHost,
		conf_influxdb: conf_influxdb {
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
	dbname := config.conf_influxdb.database
	retentionPolicy := config.conf_influxdb.retentionPolicy
	measurement := config.conf_influxdb.measurement
	// write data to database
	bp, err := client.NewBatchPoints(client.BatchPointsConfig {
		Database: dbname,
		Precision: "s",
		RetentionPolicy: retentionPolicy,
	})

	handleError(err)
	
	data := map[string]interface{}{
		"Signal": WiFi.signal, 
	}
	
	tags := map[string]string {
		"SSID": WiFi.SSID,
		"MAC Address": WiFi.MAC,
		"Frequency": WiFi.freq,
	}

	pt,err := client.NewPoint(measurement, tags, data, time.Now())
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
	} ()
	/*Start Program
	*/
	//Load config file
	config := loadConfig()
	dbaddr := config.conf_influxdb.host + ":" + config.conf_influxdb.port
	dbusername := config.conf_influxdb.username
	dbpassword := config.conf_influxdb.password
	//Create a new HTTP client sending request to InfluxDB
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:		dbaddr,
		Username:	dbusername, 	
		Password:	dbpassword,

	})

	handleError(err)

	for {
		WiFiList := ScanWiFi()
		for _, WiFi := range WiFiList {
			log.Println(WiFi)
			writeInfluxDB(clnt,config,WiFi)
		}
	}
}