package docker

import (
	"log"
	"regexp"
)

var BRACKET_PREFIX_REGEX = regexp.MustCompile(`\[([^]]*)\]\s(.*)?`)

var COMBINED_LOG_FORMAT_REGEX = regexp.MustCompile(
	`^([0-9.]+)\s` +
		`([w.-]+)\s` +
		`([\w.-]+)\s` +
		`\[([^]]+)\]\s` +
		`"((?:[^"]|")+)"\s` +
		`(\d{3})\s` +
		`(\d+|-)\s` +
		`"((?:[^"])+)"\s` +
		`"((?:[^"]|")+)"`)

func main() {
	m := map[string]interface{}{}
	augmentMapWithParsedMessage(m,
		`[httpd] 172.17.0.1 - admin [14/Nov/2017:18:07:37 +0000] "POST /write?consistency=&db=cadvisor&precision=&rp= HTTP/1.1" 204 0 "-" "cAdvisor/v0.25.0" b2a56f99-c966-11e7-9094-000000000000 18084`)
	log.Printf("map: %v", m)
}

func augmentMapWithParsedMessage(m map[string]interface{}, message string) {
	match := BRACKET_PREFIX_REGEX.FindStringSubmatch(message)
	if match != nil {
		m["prefix"] = match[1]
		message = match[2]
	}

	match = COMBINED_LOG_FORMAT_REGEX.FindStringSubmatch(message)
	if match != nil {
		m["ip_address"] = match[1]
		m["identity"] = match[2]
		m["userid"] = match[3]
		m["time"] = match[4]
		m["request"] = match[5]
		m["status_code"] = match[6]
		m["num_bytes"] = match[7]
		m["referer"] = match[8]
		m["user_agent"] = match[9]
	} else {
		m["message"] = message
	}
}
