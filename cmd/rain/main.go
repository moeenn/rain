package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"rain/internal/cli"
	"rain/internal/collection"
	"rain/internal/colors"
	"syscall"
	"time"
)

func performRequest(selectedRequest *collection.RequestEntry, timeout time.Duration) error {
	fmt.Printf("%sSending Request %s", colors.BLUE, colors.RESET)

	start := time.Now()
	resp, statusCode, err := selectedRequest.Do(
		collection.RequestArgs{
			Timeout: timeout,
		},
	)

	elapsed := time.Since(start)
	fmt.Print("\033[2J\033[H") // clear screen and move cursor to start.
	fmt.Printf("%sStatus = %s%d\t%sElapsed = %s%s\n%s", colors.BLACK, colors.YELLOW,
		statusCode, colors.BLACK, colors.YELLOW, elapsed, colors.RESET)

	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, resp, "", "  ")
	if err != nil {
		fmt.Printf("\n%s\n\n", resp)
		return nil
	}

	fmt.Printf("\n%s\n\n", prettyJson.String())
	return nil
}

func run(args []string) error {
	flags, err := cli.GetFlags(args)
	if err != nil {
		return err
	}

	// handle keyboard interrupt.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		_ = <-signals
		fmt.Printf("%s\nExiting...%s\n", colors.BLACK, colors.RESET)
		os.Exit(0)
	}()

	openedCollection, err := collection.Load(*flags.Collection, flags.Env)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	requests := openedCollection.ListRequests()
	fmt.Printf("%sSelect Request:%s\n", colors.BLUE, colors.RESET)
	for i, r := range requests {
		fmt.Printf("%2d. %s\n", i+1, r)
	}

	var selection int
	fmt.Printf("\n%sSelection: %s", colors.BLUE, colors.RESET)
	if _, err := fmt.Scanln(&selection); err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if selection < 1 || selection > len(requests) {
		return fmt.Errorf("invalid selection")
	}

	selectedRequest := openedCollection.Requests[selection-1]
	return performRequest(selectedRequest, *flags.Timeout)
}

func main() {
	args := os.Args
	if err := run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%serror: %s.%s\n", colors.RED, err.Error(), colors.RESET)
		os.Exit(1)
	}
}
