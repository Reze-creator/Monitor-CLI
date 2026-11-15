package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// WriteReport prints a readable summary and response-time distribution.
func WriteReport(output io.Writer, results []Result) {
	ordered := append([]Result(nil), results...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Target.Name < ordered[j].Target.Name })

	successes := 0
	var total time.Duration
	var slowest Result
	for i, result := range ordered {
		total += result.Duration
		if result.Success {
			successes++
		}
		if i == 0 || result.Duration > slowest.Duration {
			slowest = result
		}
	}

	fmt.Fprintf(output, "Health check report\n===================\n")
	fmt.Fprintf(output, "Targets: %d | Success: %d | Failure: %d", len(results), successes, len(results)-successes)
	if len(results) > 0 {
		fmt.Fprintf(output, " | Success rate: %.1f%%", float64(successes)*100/float64(len(results)))
	}
	fmt.Fprintln(output)
	fmt.Fprintln(output, "")
	fmt.Fprintf(output, "%-22s %-5s %-7s %-10s %-8s %s\n", "NAME", "TYPE", "RESULT", "DURATION", "TRIES", "DETAIL")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")
	for _, result := range ordered {
		status := "FAIL"
		if result.Success {
			status = "OK"
		}
		fmt.Fprintf(output, "%-22s %-5s %-7s %-10s %-8d %s\n", result.Target.Name, result.Target.Type, status, result.Duration.Round(time.Millisecond), result.Attempts, result.Message)
	}

	if len(results) == 0 {
		return
	}
	average := total / time.Duration(len(results))
	under100ms, under500ms, under1s, over1s := 0, 0, 0, 0
	for _, result := range results {
		switch {
		case result.Duration < 100*time.Millisecond:
			under100ms++
		case result.Duration < 500*time.Millisecond:
			under500ms++
		case result.Duration < time.Second:
			under1s++
		default:
			over1s++
		}
	}
	fmt.Fprintf(output, "\nResponse times: avg %s | min %s | max %s\n", average.Round(time.Millisecond), minimumDuration(results).Round(time.Millisecond), maximumDuration(results).Round(time.Millisecond))
	fmt.Fprintf(output, "Distribution: <100ms %d | 100-500ms %d | 500ms-1s %d | >=1s %d\n", under100ms, under500ms, under1s, over1s)
	fmt.Fprintf(output, "Slowest target: %s (%s)\n", slowest.Target.Name, slowest.Duration.Round(time.Millisecond))
}

func minimumDuration(results []Result) time.Duration {
	minimum := results[0].Duration
	for _, result := range results[1:] {
		if result.Duration < minimum {
			minimum = result.Duration
		}
	}
	return minimum
}

func maximumDuration(results []Result) time.Duration {
	maximum := results[0].Duration
	for _, result := range results[1:] {
		if result.Duration > maximum {
			maximum = result.Duration
		}
	}
	return maximum
}

// SaveReport writes the report to a timestamped log file in the working directory.
func SaveReport(results []Result, now time.Time) (string, error) {
	name := "monitor-log-" + now.Format("20060102150405") + ".log"
	file, err := os.Create(name)
	if err != nil {
		return "", fmt.Errorf("create report: %w", err)
	}
	defer file.Close()
	WriteReport(file, results)
	return name, nil
}
