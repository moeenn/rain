package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"rain/internal/collection"
	"rain/internal/colors"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
)

/*
TODO: implement flags.
- dump sample request collection
- env file path (optional)
- collection file path (optional)
- request timeout
- dump output to file path (optional)

TODO: implement per request based vars. Priority of reading vars
1st. Request scoped vars.
2nd. Collection vars.
3rd. Env.
*/

const (
	DEFAULT_ENV_FILEPATH        string = ".env"
	DEFAULT_COLLECTION_FILEPATH string = "collection.toml"
	DEFAULT_REQUEST_TIMEOUT     int    = 60
)

type CliArgs struct {
	Dump       *string
	Env        *string
	Collection *string
	Timeout    *time.Duration
}

func defaultCliArgs() CliArgs {
	cl := DEFAULT_COLLECTION_FILEPATH
	tout := time.Second * time.Duration(DEFAULT_REQUEST_TIMEOUT)

	return CliArgs{
		Dump:       nil,
		Env:        nil,
		Collection: &cl,
		Timeout:    &tout,
	}
}

func parseFlags(args []string) CliArgs {
	dump := flag.String("dump", "", "Dump output to the provided file path")
	env := flag.String("env", "", "Load provided environment file")
	coll := flag.String("collection", DEFAULT_COLLECTION_FILEPATH, "Load provided requests collection file")
	tout := flag.Int("timeout", DEFAULT_REQUEST_TIMEOUT, "Request timeout (seconds)")
	flag.Parse()

	timeout := time.Second * time.Duration(DEFAULT_REQUEST_TIMEOUT)
	if tout != nil {
		timeout = time.Second * time.Duration(*tout)
	}

	cliArgs := CliArgs{
		Dump:       dump,
		Env:        env,
		Collection: coll,
		Timeout:    &timeout,
	}

	if *cliArgs.Dump == "" {
		cliArgs.Dump = nil
	}

	if *cliArgs.Env == "" {
		cliArgs.Env = nil
	}

	return cliArgs
}

func run(args []string) error {
	cliArgs := defaultCliArgs()
	if len(args) > 0 {
		if args[1] == "init" {
			sampleCollection := collection.NewSampleCollection()
			encoded, err := toml.Marshal(sampleCollection)
			if err != nil {
				return fmt.Errorf("failed to encoded sample collection: %w", err)
			}

			if err := os.WriteFile(DEFAULT_COLLECTION_FILEPATH, encoded, 0644); err != nil {
				return fmt.Errorf("failed to write sample collection file: %w", err)
			}

			return nil
		} else {
			cliArgs = parseFlags(args)
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		_ = <-signals
		fmt.Printf("%s\nExiting...%s\n", colors.BLACK, colors.RESET)
		os.Exit(0)
	}()

	openedCollection, err := collection.Load(*cliArgs.Collection, *cliArgs.Env)
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
	fmt.Printf("%sSending Request %s", colors.BLUE, colors.RESET)

	start := time.Now()
	resp, statusCode, err := selectedRequest.Do(
		collection.RequestArgs{
			Timeout: *cliArgs.Timeout,
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

func main() {
	args := os.Args
	if err := run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%serror: %s.%s\n", colors.RED, err.Error(), colors.RESET)
		os.Exit(1)
	}
}
