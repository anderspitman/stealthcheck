package main

import (
	"net/http"
        "log"
        "encoding/json"
        "io/ioutil"
        "os/exec"
        "time"
        "flag"
        "path"
)

type ChecksConfig struct {
        Checks []*CheckConfig `json:"checks"`
}

type CheckConfig struct {
        IntervalMS int64 `json:"interval_ms"`
        CheckCommand string `json:"check_command"`
        SuccessCommand string `json:"success_command"`
        FailCommand string `json:"fail_command"`
}

func main() {
        log.Println("Starting up")
        port := flag.String("port", "8484", "Port")
        dir := flag.String("dir", "./", "Directory")
        flag.Parse()

        checksFilePath := path.Join(*dir, "checks.json")
        checksFile, err := ioutil.ReadFile(checksFilePath)
        if err != nil {
                log.Fatal(err)
                return
        }

        config := &ChecksConfig{}
        err = json.Unmarshal(checksFile, config)
        if err != nil {
                log.Fatal(err)
                return
        }

        validateConfig(config)

        for _, check := range config.Checks {
                go startJob(check)
        }

	handler := func(w http.ResponseWriter, r *http.Request) {
                // returns HTTP 200 by default
        }

        err = http.ListenAndServe(":"+*port, http.HandlerFunc(handler))
	if err != nil {
		log.Println(err)
	}
}


func startJob(check *CheckConfig) {
        for {
                command := exec.Command("sh", "-c", check.CheckCommand)
                _, err := command.Output()
                if err != nil {
                        log.Println("Check failed:", check.CheckCommand)
                        go runFailCommand(check)
                }

                time.Sleep(time.Duration(check.IntervalMS) * time.Millisecond)
        }
}

func validateConfig(config *ChecksConfig) {
        for _, check := range config.Checks {
                if check.IntervalMS < 100 {
                        log.Fatal("Min 100 MS interval")
                }
        }
}

func runFailCommand(check *CheckConfig) {
        command := exec.Command("sh", "-c", check.FailCommand)
        err := command.Start()
        if err != nil {
                log.Println("Fail command failed:", check.FailCommand)
        }
}
