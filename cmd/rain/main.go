package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"rain/internal/collection"
	"rain/internal/colors"
	"syscall"
)

const (
	DEFAULT_ENV_FILEPATH        string = ".env"
	DEFAULT_COLLECTION_FILEPATH string = "collection.toml"
)

func run() error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		_ = <-signals
		fmt.Printf("%s\nExiting...%s\n", colors.BLACK, colors.RESET)
		os.Exit(0)
	}()

	collection, err := collection.Load(DEFAULT_COLLECTION_FILEPATH, DEFAULT_ENV_FILEPATH)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	/*
		requests := collection.ListRequests()
		for {
			fmt.Printf("%sSelect Request:%s\n", colors.BLUE, colors.RESET)
			for i, r := range requests {
				fmt.Printf("%2d. %s\n", i+1, r)
			}

			var selection int
			fmt.Printf("\n%sSelection: %s", colors.BLUE, colors.RESET)
			if _, err := fmt.Scanln(&selection); err != nil {
				fmt.Printf("%sfailed to read input: %s.%s", colors.RED, err.Error(), colors.RESET)
				continue
			}

			if selection < 1 || selection > len(requests) {
				fmt.Printf("%sInvalid selection: Please try again.%s\n", colors.RED, colors.RESET)
				continue
			}
		}
	*/

	e, _ := json.Marshal(collection)
	fmt.Printf("%s\n", e)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}
