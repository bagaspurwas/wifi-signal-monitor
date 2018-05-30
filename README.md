# wifi-signal-monitor

## Getting Started

This program is written in specific scenario that is to monitor wifi signal quality in office. The idea is deploying SBC around the office, connected to a network and send retrieved data to a database server (influxDB). The data then can be visualized using Grafana or whatever, this helps office position their Wireless Device efficiently.
Of course other use cases are possible.

### Requirement

- InfluxDB (v1.2) to store data, please refer to official website or github page
```
```
- Golang
Golang is mandatory to build this program.
```
github.com/golang/go
```

### Installing
```
to be added
```

### Example Usage
I quickly deploy influxdb and grafana on top of docker to see this program in action. Link to github: https://github.com/philhawthorne/docker-influxdb-grafana

Get neccessary libs and run the program
```
go get <libs>
go run main.go
```

Add influxDB data source and a graph on Grafana dashboard. Provide the following query:
```
SELECT "Signal" FROM "onehour"."SIGNALSTR" WHERE $timeFilter GROUP BY "MAC Address", "SSID"
```
Please note that retention policy, measurement unit can vary, please edit to your liking in configuration file.

This is the result displayed on Grafana
![alt text](http://i63.tinypic.com/29d6yr5.png)
