package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/tombuildsstuff/huawei-e5573-mifi-sdk-go/mifi"
)

var versionNumber = "0.0.4"
var e5573URL = flag.String("e5573-url", "http://192.168.8.1", "The endpoint of the E5573 device")
var showVersion = flag.Bool("version", false, "Show the Application Version")
var verbose = flag.Bool("verbose", false, "Produce verbose output")
var showHelp = flag.Bool("help", false, "Displays this message")
var influxURL = flag.String("influxdb-url", "http://localhost:8086", "URL to the InfluxDB where samples are sent to")
var influxDatabase = flag.String("influxdb-database", "", "InfluxDB database name where samples are written to")
var influxUsername = flag.String("influxdb-user", "", "InfluxDB user name that can write samples to the given database.")

func main() {
	flag.Parse()

	if *showHelp {
		flag.Usage()
		return
	}

	if *showVersion {
		fmt.Printf("v%s\n", versionNumber)
		return
	}

	if *e5573URL == "" {
		fmt.Fprintf(os.Stderr, "Error: missing mandatory E5573 device URL.\n")
		os.Exit(1)
	}

	mifi := mifi.Mifi{Endpoint: *e5573URL}

	status, err := mifi.CurrentStatus()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	battery := float64(status.CurrentBatteryPercentage) / 100.0
	wifiUsers := status.NumberOfUsersConnectedToWifi

	networkSettings, err := mifi.NetworkSettings()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	networkMode := networkSettings.NetworkMode()
	networkSignal := float64(status.CurrentSignalBars) / float64(status.MaxSignalBars)

	traffic, err := mifi.TrafficStatistics()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	connectionTime := traffic.SecondsConnectedToNetwork
	downloadedMiB := traffic.DownloadedMB
	uploadedMiB := traffic.UploadedMB

	if *verbose == true {
		fmt.Println("Version: ", versionNumber)
		fmt.Println("Timestamp: ", time.Now().Format(time.RFC3339))
		fmt.Printf("Battery: %.2f\n", battery)
		fmt.Printf("WiFi Users: %d\n", wifiUsers)
		fmt.Printf("Network Mode: %s\n", networkMode)
		fmt.Printf("Network Signal: %.2f\n", networkSignal)
		fmt.Printf("Connection Time: %d s\n", connectionTime)
		fmt.Printf("Downloaded: %.2f MiB\n", downloadedMiB)
		fmt.Printf("Uploaded %.2f MiB\n", uploadedMiB)
	}

	if *influxDatabase != "" {
		influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
			Addr:     *influxURL,
			Username: *influxUsername,
			Password: os.Getenv("INFLUXDB_PASSWORD"),
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not connect to InfluxDB at %v: %v\n", *influxURL, err)
			os.Exit(1)
		}

		tags := map[string]string{"hostname": hostName()}

		bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
			Database:  *influxDatabase,
			Precision: "s",
		})

		if err != nil {
			fmt.Println("Error creating new data points for InfluxDB: ", err)
			os.Exit(1)
		}

		addSystemPoint(bp, battery, wifiUsers, tags)
		addNetworkPoint(bp, networkMode, networkSignal, tags)
		addTrafficPoint(bp, connectionTime, downloadedMiB, uploadedMiB, tags)

		err = influxClient.Write(bp)

		if err != nil {
			fmt.Println("Error writing data points to InfluxDB: ", err)
			os.Exit(1)
		}
	}
}

func addSystemPoint(bp influx.BatchPoints, battery float64, wifiUsers int, tags map[string]string) error {
	fields := map[string]interface{}{"battery": battery, "wifi_users": wifiUsers}
	return addPoint(bp, "system", tags, fields)
}

func addNetworkPoint(bp influx.BatchPoints, mode string, signal float64, tags map[string]string) error {
	fields := map[string]interface{}{"mode": mode, "signal": signal}
	return addPoint(bp, "network", tags, fields)
}

func addTrafficPoint(bp influx.BatchPoints, connectionTime int, downloadedMiB, uploadedMiB float32, tags map[string]string) error {
	fields := map[string]interface{}{"connection_time": connectionTime, "downloaded": downloadedMiB, "uploaded": uploadedMiB}
	return addPoint(bp, "traffic", tags, fields)
}

func addPoint(bp influx.BatchPoints, name string, tags map[string]string, fields map[string]interface{}) error {
	pt, err := influx.NewPoint(name, tags, fields)

	if err != nil {
		return err
	}

	bp.AddPoint(pt)

	return nil
}

func hostName() string {
	hostName, err := os.Hostname()

	if err != nil {
		os.Stderr.WriteString("Warning: Could not determine hostname; using 'unknown'.\n")
		hostName = "unknown"
	}

	return hostName
}
