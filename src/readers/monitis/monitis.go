package monitis

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/danielstutzman/go-monitis"
	"github.com/danielstutzman/sync-log-files-to-db/src/log"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/influxdb"
	"github.com/danielstutzman/sync-log-files-to-db/src/writers/postgres"
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

	var postgresConn *postgres.PostgresConnection
	if config.Postgresql != nil {
		postgresConn = postgres.NewPostgresConnection(config.Postgresql, configPath)
		postgresConn.CreateMonitisResultsTable()
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
			pollMonitisMonitorForever(&monitorCopy, auth, influxdbConn, postgresConn)
		}(monitor)
	}
	if len(monitors) > 0 {
		select {} // block forever
	}
}

func pollMonitisMonitorForever(monitor *monitis.ExternalMonitor,
	auth *monitis.Auth,
	influxdbConn *influxdb.InfluxdbConnection,
	postgresConn *postgres.PostgresConnection) {

	var earliestInfluxDate time.Time
	if influxdbConn != nil {
		earliestInfluxDate = midnightBefore(influxdbConn.QueryForLastTimestampForTag(
			"response_millis", "monitor_name", monitor.Name))
		log.Infow("Most recent timestamp in InfluxDB",
			"timestamp", earliestInfluxDate, "monitor_name", monitor.Name)
	}

	var earliestPostgresDate time.Time
	if postgresConn != nil {
		earliestPostgresDate = postgresConn.QueryForLastTimestamp(
			fmt.Sprintf("monitor_name = %s", postgres.QuoteString(monitor.Name)))
		log.Infow("Most recent timestamp in Postgresql",
			"timestamp", earliestPostgresDate, "monitor_name", monitor.Name)
	}

	var retrieveDate time.Time
	if !earliestInfluxDate.IsZero() && !earliestPostgresDate.IsZero() {
		if earliestInfluxDate.Before(earliestPostgresDate) {
			retrieveDate = earliestPostgresDate
		} else {
			retrieveDate = earliestInfluxDate
		}
	} else if !earliestInfluxDate.IsZero() {
		retrieveDate = earliestInfluxDate
	} else if !earliestPostgresDate.IsZero() {
		retrieveDate = earliestPostgresDate
	} else {
		// Since there's no existing data in InfluxDB/Postgresql,
		// find out the monitor's start date and start scraping from then
		info, err := auth.GetExternalMonitorInfo(strconv.Itoa(monitor.Id), nil)
		if err != nil {
			log.Fatalw("Error from GetExternalMonitorInfo",
				"err", err, "monitor_name", monitor.Name)
		}
		retrieveDate = info.StartDateParsed
	}

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
		log.Infow("Calling GetExternalResults...",
			"date", retrieveDate.Format("2006-01-02"),
			"monitor_name", monitor.Name)
		results, err := auth.GetExternalResults(strconv.Itoa(monitor.Id),
			&monitis.GetExternalResultsOptions{
				Year:  monitis.Int(retrieveDate.Year()),
				Month: monitis.Int(int(retrieveDate.Month())),
				Day:   monitis.Int(retrieveDate.Day()),
			})
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
			if influxdbConn != nil {
				influxdbConn.InsertMaps(INFLUXDB_TAGS_SET, maps)
			}
			if postgresConn != nil {
				postgresConn.InsertMaps(maps)
			}
		}

		if retrieveDate.Before(midnightBefore(time.Now().UTC())) {
			retrieveDate = retrieveDate.AddDate(0, 0, 1)

			// Throttle API a little
			time.Sleep(time.Second)
		} else {
			log.Infow("Wait until next point posted", "minutes", minIntervalMinutes)
			time.Sleep(time.Duration(minIntervalMinutes) * time.Minute)
		}
	}
}

// Rounds time down to the day
func midnightBefore(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
