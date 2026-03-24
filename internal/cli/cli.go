package cli

import (
	"flag"
	"fmt"
	"os"
	"rain/internal/collection"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultEnvFilepath        string = ".env"
	defaultCollectionFilepath string = "collection.yml"
	defaultRequestTimeout     int    = 60
)

type CliArgs struct {
	Dump       *string
	Env        *string
	Collection *string
	Timeout    *time.Duration
}

func defaultCliArgs() CliArgs {
	cl := defaultCollectionFilepath
	tout := time.Second * time.Duration(defaultRequestTimeout)

	return CliArgs{
		Dump:       nil,
		Env:        nil,
		Collection: &cl,
		Timeout:    &tout,
	}
}

func parseFlags() CliArgs {
	dump := flag.String("dump", "", "Dump output to the provided file path")
	env := flag.String("env", "", "Load provided environment file")
	coll := flag.String("collection", defaultCollectionFilepath, "Load provided requests collection file")
	tout := flag.Int("timeout", defaultRequestTimeout, "Request timeout (seconds)")
	flag.Parse()

	timeout := time.Second * time.Duration(defaultRequestTimeout)
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

func dumpDefaultCollection() error {
	sampleCollection := collection.NewSampleCollection()
	encoded, err := yaml.Marshal(sampleCollection)
	if err != nil {
		return fmt.Errorf("failed to encoded sample collection: %w", err)
	}

	if err := os.WriteFile(defaultCollectionFilepath, encoded, 0644); err != nil {
		return fmt.Errorf("failed to write sample collection file: %w", err)
	}

	return nil
}

func GetFlags(args []string) (*CliArgs, error) {
	cliArgs := defaultCliArgs()
	if len(args) > 1 {
		command := args[1]
		switch command {
		case "init":
			if err := dumpDefaultCollection(); err != nil {
				return nil, err
			}

		default:
			cliArgs = parseFlags()
		}
	}

	return &cliArgs, nil
}
