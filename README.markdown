`Agent E5573` collects some stats from the [Huawei E5573](https://en.wikipedia.org/wiki/Huawei_E5#Huawei_E5573) wireless router ("MiFi") and publishes them to an InfluxDB instance.

# TODO

* Monitor whether the battery is being charged with `/response/BatteryStatus` (it is 0 or 1)

  ```command
  $ curl 'http://192.168.8.1/api/monitoring/status' \
    -H 'Connection: keep-alive' \
    -H 'Accept: */*' \
    -H 'DNT: 1' \
    -H 'X-Requested-With: XMLHttpRequest' \
    -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36' \
    -H 'Referer: http://192.168.8.1/html/statistic.html' \
    -H 'Accept-Language: en-US,en;q=0.9,de;q=0.8' \
    -H 'Cookie: SessionID=9hL2skrdEu7T+3+xV3dUYcvDjjVmeTI8qaTMQ54XUUEdcwOjp1NjAbcuwB2gOdpLO1//f1pMJ4exh2IREr4IZe8atxU2WnQVvSnjYeG+XQBWqA1wjXPHTNWuIBZ/+E/y' \
    --compressed \
    --insecure
  <?xml version="1.0" encoding="UTF-8"?>
  <response>
    <ConnectionStatus>901</ConnectionStatus>
    <WifiConnectionStatus>902</WifiConnectionStatus>
    <SignalStrength></SignalStrength>
    <SignalIcon>5</SignalIcon>
    <CurrentNetworkType>9</CurrentNetworkType>
    <CurrentServiceDomain>3</CurrentServiceDomain>
    <RoamingStatus>0</RoamingStatus>
    <BatteryStatus>0</BatteryStatus>
    <BatteryLevel>4</BatteryLevel>
    <BatteryPercent>100</BatteryPercent>
    <simlockStatus>0</simlockStatus>
    <PrimaryDns>139.7.30.126</PrimaryDns>
    <SecondaryDns>139.7.30.125</SecondaryDns>
    <PrimaryIPv6Dns></PrimaryIPv6Dns>
    <SecondaryIPv6Dns></SecondaryIPv6Dns>
    <CurrentWifiUser>1</CurrentWifiUser>
    <TotalWifiUser>16</TotalWifiUser>
    <currenttotalwifiuser>16</currenttotalwifiuser>
    <ServiceStatus>2</ServiceStatus>
    <SimStatus>1</SimStatus>
    <WifiStatus>1</WifiStatus>
    <CurrentNetworkTypeEx>46</CurrentNetworkTypeEx>
    <WanPolicy>0</WanPolicy>
    <maxsignal>5</maxsignal>
    <wifiindooronly>0</wifiindooronly>
    <wififrequence>0</wififrequence>
    <classify>mobile-wifi</classify>
    <flymode>0</flymode>
    <cellroam>1</cellroam>
    <ltecastatus>0</ltecastatus>
  </response>
  ```

* Clear stats with a POST to: http://192.168.8.1/api/monitoring/clear-traffic

  Needs authentication, which we don't have yet. Good info here: https://github.com/zikusooka/query_huawei_wifi_router/blob/master/query_huawei_wifi_router.sh

  ```command
  curl 'http://192.168.8.1/api/monitoring/clear-traffic' \
    -H 'Connection: keep-alive' \
    -H 'Accept: */*' \
    -H 'DNT: 1' \
    -H 'X-Requested-With: XMLHttpRequest' \
    -H '__RequestVerificationToken: Vccg5QdWEvTYgNyoX9w1/p5IPR3cZlOp' \
    -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36' \
    -H 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' \
    -H 'Origin: http://192.168.8.1' \
    -H 'Referer: http://192.168.8.1/html/statistic.html' \
    -H 'Accept-Language: en-US,en;q=0.9,de;q=0.8' \
    -H 'Cookie: SessionID=kN9uW5X0xEIvvCKUKIiqCuueM3CWwSDs08Ip3SHxIK9t06OSw577OOv4N4vCPHQUmFPd0R18kQaP2ryEDJtE0bUSvw4HP3DqDJ00n0kcsFFCvXkTah8rBfybYSoQ0OUi' \
    --data-raw '<?xml version="1.0" encoding="UTF-8"?><request><ClearTraffic>1</ClearTraffic></request>' \
    --compressed \
    --insecure
  ```

* Extract deployment and pipeline from tasmota-sensor-bridge

# Synopsis

```command
$ export INFLUXDB_PASSWORD=s3cret
$ agent-e5573 \
  --e5573-url http://192.168.8.1 \
  --influxdb-url https://alice:s3cret@influxdb.example.com/e5573
```

Run it with `--verbose` and it will print its the stats to `STDOUT`:

```command
$ agent-e5573 \
  --verbose \
  --e5573-url http://192.168.8.1 \
  --influxdb-url https://alice:s3cret@influxdb.example.com/e5573
Timestamp: 2020-05-11T19:12:21+02:00
Battery: 1.00
WiFi Users: 1
Network Mode: 4G/LTE Enabled
Network Signal: 0.20
Connection Time: 49764 s
Downloaded: 93.64 MiB
Uploaded 98.70 MiB
```

Run it without an InfluxDB URL, and it will just print the stats without attempting to write to an InfluxDB.

# System Design Choices

We want to publish the stats in regular intervals. The following choices come to mind:

1. systemd timers
1. cron
1. Keep it running and maintain a loop that publishes directly
1. New [Telegraf plugin](https://www.influxdata.com/blog/telegraf-go-collection-agent/)
1. Provide data to the [Telegraf exec plugin](https://community.influxdata.com/t/data-collection-question-best-way-to-feed-from-a-stats-catcher/11964)
1. Act as Prometheus exporter and use the [Telegraf Prometheus plugin](https://community.influxdata.com/t/own-telegraf-plugin-need-to-scrape-metrics-from-prometheus-clients/11878)

Since we do not publish more than once a minute, the granularity of systemd timers or cron seems to be sufficient. And it helps keeping things simple.

# Development

Regular deployment is done with Ansible. For fast iteration, the following approach can be used:

1. Set the following environment variables:

    ```command
    $ export AGENT_E5573_HOST=pi.example.com # where the agent will be running
    $ export INFLUXDB_URL=https://alice:s3cret@influxdb.example.com/e5573 # where the agent pushes data to
    ```

1. Optionally, if your target system is a Raspberry Pi <= 3, set the following variables on the build machine:

    ```command
    $ export GOOS=linux
    $ export GOARCH=arm
    $ export GOARM=7
    ```

1. Run the `setup` script once:

    ```command
    $ scripts/setup
    ```

1. Run this combination of scripts in order to `build`, `deploy` and `run` the agent:

    ```command
    $ scripts/build && scripts/deploy && scripts/run
    ```

    If you want to iterate quickly and build, deploy and run on each save, run `scripts/watch`.

1. Optionally, improve the code and supply a pull request.
