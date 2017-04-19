package lambda_deployer

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func findGitSha1(path string) string {
	gitCommand := []string{"git", "rev-parse", "--short", "HEAD"}
	log.Printf("Running %s...", strings.Join(gitCommand, " "))
	cmd := exec.Command(gitCommand[0], gitCommand[1:]...)
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error from StdoutPipe: %s", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Error from cmd.Start: %s", err)
	}
	stdoutBytes, err := ioutil.ReadAll(stdoutReader)
	if err != nil {
		log.Fatalf("Error from ReadAll: %s", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Error from cmd.Wait: %s", err)
	}
	return strings.TrimSpace(string(stdoutBytes))
}

// Returns created zip file
func zip() string {
	gitSha1 := findGitSha1(".")
	zipPath := "build/" + gitSha1 + ".zip"
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		zipCommand := []string{"/bin/bash", "-c", `mkdir -p build &&
			cp src/NodeWrapper.js build &&
			rm -f $GOPATH/bin/linux_amd64/deployed build/deployed &&
			GOOS=linux GOARCH=amd64 go install github.com/danielstutzman/sync-cloudfront-logs-to-bigquery/src/...
			cp $GOPATH/bin/linux_amd64/deployed build &&
			cd build &&
			zip -r -q ../` + zipPath + ` NodeWrapper.js deployed &&
			rm -f build/deployed`}
		log.Printf("Running %s...", strings.Join(zipCommand, " "))
		cmd := exec.Command(zipCommand[0], zipCommand[1:]...)
		stdoutReader, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("Error from StdoutPipe: %s", err)
		}
		stderrReader, err := cmd.StderrPipe()
		if err != nil {
			log.Fatalf("Error from StderrPipe: %s", err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatalf("Error from cmd.Start: %s", err)
		}
		stdoutScanner := bufio.NewScanner(stdoutReader)
		stderrScanner := bufio.NewScanner(stderrReader)
		for stdoutScanner.Scan() {
			log.Printf("Output from zipCommand: %s", stdoutScanner.Text())
		}
		for stderrScanner.Scan() {
			log.Printf("Output from zipCommand: %s", stderrScanner.Text())
		}
		if err := cmd.Wait(); err != nil {
			log.Fatalf("Error from cmd.Wait: %s", err)
		}
	}
	// TODO: check porcelain for uncommitted
	return zipPath
}
