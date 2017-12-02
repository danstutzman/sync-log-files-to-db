package monitis

import (
	"strconv"
	"strings"
	"time"

	"github.com/danielstutzman/go-monitis"
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/storage/influxdb"
)

var INFLUXDB_TAGS_SET = map[string]bool{
	"monitor_name":  true,
	"location_name": true,
}

func PollMonitisForever(config *Options, configPath string) {
	var influxdbConn *influxdb.InfluxdbConnection
	if config.InfluxDb != nil {
		influxdbConn = influxdb.NewInfluxdbConnection(config.InfluxDb, configPath)
		influxdbConn.CreateDatabase()
	}

	auth, err := monitis.GetAuthToken(config.ApiKey, config.SecretKey)
	if err != nil {
		log.Fatalw("Error from GetAuthToken", "err", err)
	}

	monitors, err := auth.GetExternalMonitors()
	if err != nil {
		log.Fatalw("Error from GetExternalMonitors", "err", err)
	}

	for _, monitor := range monitors {
		go func(monitorCopy monitis.ExternalMonitor) {
			pollMonitisMonitorForever(&monitorCopy, auth, influxdbConn)
		}(monitor)
	}
	if len(monitors) > 0 {
		select {} // block forever
	}
}

func pollMonitisMonitorForever(monitor *monitis.ExternalMonitor,
	auth *monitis.Auth, influxdbConn *influxdb.InfluxdbConnection) {

	minIntervalMinutes := 9999
	for _, intervalString := range strings.Split(monitor.Intervals, ",") {
		interval, err := strconv.Atoi(intervalString)
		if err != nil {
			log.Fatalw("Error from Itoa", "input", intervalString)
		}
		if interval < minIntervalMinutes {
			minIntervalMinutes = interval
		}
	}

	for {
		results, err := auth.GetExternalResults(strconv.Itoa(monitor.Id))
		if err != nil {
			log.Fatalw("Error from GetExternalResults",
				"err", err, "monitor_name", monitor.Name)
		}

		maps := []map[string]interface{}{}
		for _, result := range results {
			for _, point := range result.Points {
				m := map[string]interface{}{
					"monitor_name":    monitor.Name,
					"location_name":   result.LocationName,
					"timestamp":       point.Timestamp,
					"response_millis": point.Duration,
					"was_okay":        point.WasOkay,
				}
				maps = append(maps, m)
			}
		}

		if len(maps) > 0 {
			influxdbConn.InsertMaps(INFLUXDB_TAGS_SET, maps)
		}

		log.Infow("Wait until next point posted", "minutes", minIntervalMinutes)
		time.Sleep(time.Duration(minIntervalMinutes) * time.Minute)
	}
}
