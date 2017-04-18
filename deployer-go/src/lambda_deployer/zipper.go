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
func zip(pathToDeploy string) string {
	gitSha1 := findGitSha1(pathToDeploy)
	zipPath := "../deployed/build/" + gitSha1 + ".zip"
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		zipCommand := []string{"/bin/bash", "-c", `cd ` + pathToDeploy + ` &&
			npm install &&
			mkdir -p build &&
			cat src/CreateThumbnail.js > CreateThumbnail.js &&
			zip -r -q ../deployer/` + zipPath + ` CreateThumbnail.js node_modules &&
			rm -f CreateThumbnail.js`}
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
