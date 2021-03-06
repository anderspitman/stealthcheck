stealthcheck is a dead-simple service monitoring tool in <150 lines of code
with no dependencies other than Go. It's so simple you probably should write
your own instead of using it.

It provides an easy way to set up health checks, automatic restarts, and email
alerts for failed checks. It also has basic logging.

Everything is based off shell commands shoved into a json config file. For each
service, you run a check command at a given interval. If the command returns
anything other than 0, run the fail command.


# Who watches the watchers?

For redundancy, simply set up 2 or more stealthcheck instances with commands to
check and restart each other. To facilitate this, HTTP requests to stealthcheck
return an empty response with 200 status.


# Example configs

Note that if a list of alert emails is specified for a specific check, none of
the alert email addresses in the main config will be alerted.

## Main config

```json
{
  "alert_emails": [
    "alice@example.com",
    "bob@example.com"
  ],
  "smtp": {
    "server": "smtp.fastmail.com",
    "port": 587,
    "sender": "someone@example.com",
    "username": "someone@example.com",
    "password": "************"
  }
}
```

## checks.json

stealthcheck instance 1:
```json
{
  "checks": [
    {
      "interval_ms": 10000,
      "check_command": "curl localhost:8486",
      "fail_command": "../stealthcheck -port 8486 -dir ../stealth2 >> ../stealth2/log.txt 2>&1 &",
      "alert_emails": [
        "someone@example.com"
      ]
    },
    {
      "interval_ms": 120000,
      "check_command": "curl service1.example.com",
      "fail_command": "ssh ubuntu@example.com ./service1"
    },
    {
      "interval_ms": 120000,
      "check_command": "./custom_checker_script.sh",
      "fail_command": "ssh ubuntu@example.com ./service2"
    }
  ]
}
```

stealthcheck instance 2
```json
{
  "checks": [
    {
      "interval_ms": 10000,
      "check_command": "curl localhost:8485",
      "fail_command": "../stealthcheck -port 8485 -dir ../stealth1 >> ../stealth1/log.txt 2>&1 &",
      "alert_emails": [
        "someone@example.com",
        "someone_else@example.com"
      ]
    }
  ]
}
```
