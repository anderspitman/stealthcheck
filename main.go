package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os/exec"
	"path"
	"time"
)

type Config struct {
	AlertEmails []string `json:"alert_emails"`
	Smtp        *SmtpConfig
}

type SmtpConfig struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Sender   string `json:"sender"`
}

type ChecksConfig struct {
	Checks []*CheckConfig `json:"checks"`
}

type CheckConfig struct {
	IntervalMS     int64    `json:"interval_ms"`
	CheckCommand   string   `json:"check_command"`
	SuccessCommand string   `json:"success_command"`
	FailCommand    string   `json:"fail_command"`
	AlertEmails    []string `json:"alert_emails"`
}

func main() {
	log.Println("Starting up")
	port := flag.String("port", "8484", "Port")
	dir := flag.String("dir", "./", "Directory")
	flag.Parse()

	configFilePath := path.Join(*dir, "config.json")
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	config := &Config{}
	err = json.Unmarshal(configFile, config)
	if err != nil {
		log.Fatal(err)
		return
	}

	checksFilePath := path.Join(*dir, "checks.json")
	checksFile, err := ioutil.ReadFile(checksFilePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	checksConfig := &ChecksConfig{}
	err = json.Unmarshal(checksFile, checksConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	validateChecks(checksConfig)

	for _, check := range checksConfig.Checks {
		go startJob(config, check)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// returns HTTP 200 by default
	}

	err = http.ListenAndServe(":"+*port, http.HandlerFunc(handler))
	if err != nil {
		log.Println(err)
	}
}

func startJob(config *Config, check *CheckConfig) {
	for {
		command := exec.Command("bash", "-c", check.CheckCommand)
		_, err := command.Output()
		if err != nil {
			log.Println("Check failed:", check.CheckCommand)
			go runFailCommand(check)

			if len(check.AlertEmails) > 0 {
				sendEmails(config.Smtp, check.AlertEmails, check.CheckCommand)
			} else if len(config.AlertEmails) > 0 {
				sendEmails(config.Smtp, config.AlertEmails, check.CheckCommand)
			}
		}

		time.Sleep(time.Duration(check.IntervalMS) * time.Millisecond)
	}
}

func validateChecks(checksConfig *ChecksConfig) {
	for _, check := range checksConfig.Checks {
		if check.IntervalMS < 100 {
			log.Fatal("Min 100 MS interval")
		}
	}
}

func sendEmails(smtpConfig *SmtpConfig, emails []string, command string) {
	for _, email := range emails {
		sendEmail(smtpConfig, email, command)
	}
}

func sendEmail(smtpConfig *SmtpConfig, email string, command string) {
	log.Println("Alerting " + email)
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Server)
	to := []string{email}
	msg := []byte(fmt.Sprintf("To: %s\r\n", email) +
		fmt.Sprintf("From: stealthcheck automated alerts <%s>\r\n", smtpConfig.Sender) +
		"Subject: stealthcheck command failed\r\n" +
		"\r\n" +
		"The following command failed: '" + command + "'\r\n")
	addr := fmt.Sprintf("%s:%d", smtpConfig.Server, smtpConfig.Port)
	err := smtp.SendMail(addr, auth, smtpConfig.Sender, to, msg)
	if err != nil {
		fmt.Println(err)
		log.Println("Sending alert email failed: " + email)
	}
}

func runFailCommand(check *CheckConfig) {
	command := exec.Command("bash", "-c", check.FailCommand)
	err := command.Run()
	if err != nil {
		log.Println("Fail command failed:", check.FailCommand)
	}
}
