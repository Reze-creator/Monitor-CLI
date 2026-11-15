package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	configPath := flag.String("config", "config.json", "path to the JSON config file")
	timeoutSeconds := flag.Int("timeout", 3, "timeout for each probe, in seconds")
	verbose := flag.Bool("v", false, "print each target result as it completes")
	flag.Parse()

	if *timeoutSeconds <= 0 {
		fmt.Fprintln(os.Stderr, "timeout must be greater than zero")
		os.Exit(2)
	}
	config, err := ReadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var onResult func(Result)
	if *verbose {
		onResult = func(result Result) {
			status := "FAIL"
			if result.Success {
				status = "OK"
			}
			fmt.Printf("[%s] %-20s %s (%s, %d attempt(s))\n", time.Now().Format("15:04:05"), result.Target.Name, status, result.Duration.Round(time.Millisecond), result.Attempts)
		}
	}

	results := RunProbes(config.Targets, time.Duration(*timeoutSeconds)*time.Second, onResult)
	WriteReport(os.Stdout, results)
	logName, err := SaveReport(results, time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("\nReport saved to %s\n", logName)
}
