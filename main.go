package main

import (
        "log"
        "encoding/json"
        "io/ioutil"
        "os"
        "os/exec"
        "os/signal"
        "sync"
        "time"
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
        checksFilePath := "checks.json"
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

        waitForCtrlC()
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
        _, err := command.Output()
        if err != nil {
                log.Println("Fail command failed:", check.FailCommand)
        }
}

func waitForCtrlC() {
        // copied from https://jjasonclark.com/waiting_for_ctrl_c_in_golang/
        var end_waiter sync.WaitGroup
        end_waiter.Add(1)
        var signal_channel chan os.Signal
        signal_channel = make(chan os.Signal, 1)
        signal.Notify(signal_channel, os.Interrupt)
        go func() {
            <-signal_channel
            end_waiter.Done()
        }()
        end_waiter.Wait()
}
