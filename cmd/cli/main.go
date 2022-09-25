package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	mer "github.com/vmw-pso/eve/mer"
)

type config struct {
	Log bool
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		logging = flags.Bool("log", false, "whether to log to file or not")
		convert = flags.String("convert", "", "filename to convert")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	c := config{
		Log: *logging,
	}

	if c.Log {
		log.Println("Logging true")
	}

	filename := *convert
	if filename != "" {
		converter := mer.NewConverter(filename)
		return converter.Convert()
	} else {
		fmt.Printf("Value of filename = %s\n", filename)
	}
	return nil
}
