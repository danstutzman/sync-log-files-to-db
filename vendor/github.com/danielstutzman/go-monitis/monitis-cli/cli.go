package main

import (
	"flag"
	"github.com/danielstutzman/go-monitis"
	"log"
)

type Config struct {
	ApiKey    string
	SecretKey string
}

func main() {
	config := Config{}
	flag.StringVar(&config.ApiKey, "apikey", "", "API key for Monitis")
	flag.StringVar(&config.SecretKey, "secretkey", "", "Secret key for Monitis")
	flag.Parse()

	if config.ApiKey == "" {
		log.Fatalf("Missing -apikey")
	}
	if config.SecretKey == "" {
		log.Fatalf("Missing -secretkey")
	}

	auth, err := monitis.GetAuthToken(config.ApiKey, config.SecretKey)
	if err != nil {
		log.Fatalf("Error from GetAuthToken: %s", err)
	}

	if false {
		alerts, err := auth.GetRecentAlerts()
		if err != nil {
			log.Fatalf("Error from GetRecentAlerts: %s", err)
		}
		log.Printf("Recent alerts: %+v", alerts)
	}

	if false {
		m := monitis.AddExternalMonitorOptions{
			Name:             monitis.String("myname"),
			DetailedTestType: monitis.Int(1),
			Tag:              monitis.String("Default"),
			LocationIds:      monitis.String("1,9"),
			Url:              monitis.String("www.example.com"),
			Type:             monitis.String("http"),
			Interval:         monitis.Int(15),
		}
		monitor, err := auth.AddExternalMonitor(&m)
		if err != nil {
			log.Fatalf("Error from AddExternalMonitor: %s", err)
		}
		log.Printf("New monitor: %+v", monitor)
	}

	if true {
		monitors, err := auth.GetExternalMonitors()
		if err != nil {
			log.Fatalf("Error from GetExternalMonitors: %s", err)
		}
		log.Printf("External monitors: %+v", monitors)
	}

	if false {
		err := auth.DeleteExternalMonitors("1385112", nil)
		if err != nil {
			log.Fatalf("Error from DeleteExternalMonitors: %s", err)
		}
	}

	if false {
		testId := "917942"
		results, err := auth.GetExternalResults(testId,
			&monitis.GetExternalResultsOptions{
				Year:  monitis.Int(2016),
				Month: monitis.Int(9),
				Day:   monitis.Int(20),
			})
		if err != nil {
			log.Fatalf("Error from GetExternalResults: %s", err)
		}
		log.Printf("External results: %+v", results)
	}

	if false {
		locations, err := auth.GetLocations()
		if err != nil {
			log.Fatalf("Error from GetLocations: %s", err)
		}
		log.Printf("Locations: %+v", locations)
	}

	if true {
		testId := "917942"
		info, err := auth.GetExternalMonitorInfo(testId, nil)
		if err != nil {
			log.Fatalf("Error from GetExternalMonitorInfo: %s", err)
		}
		log.Printf("Monitor info: %+v", info)
	}
}
