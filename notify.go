package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

func smartLog(message, urgency string, alert bool) {
	if alert {
		if err := notify(message, urgency); err != nil {
			log.Printf("failure sending notification: %s\n", err)
		}
	}

	switch urgency {
	case "critical", "normal":
		log.Fatal(message)
	default:
		log.Println(message)
	}
}

func notify(message, urgency string) error {
	var timeout time.Duration
	switch urgency {
	// cases accepted by notify-send: low, normal, critical
	case "critical":
		timeout = 5 * time.Second
	case "low":
		timeout = 2 * time.Second
	default:
		timeout = 3 * time.Second
	}

	// notify-send only accepts milliseconds
	millisec := fmt.Sprintf("%v", timeout.Seconds()*1000)

	args := []string{"-a", "Clipman", "-u", urgency, "-t", millisec, message}

	return exec.Command("notify-send", args...).Run()
}
