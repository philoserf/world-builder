// Command wbh generates a Mongoose Traveller World Builder's Handbook
// star system from a seed and emits its IISS Class 0/I Survey form.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"wbh/roller"
	"wbh/stars"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("wbh", flag.ContinueOnError)
	fs.SetOutput(stderr)
	seed := fs.Int64("seed", 0, "random seed (0 = time-based)")
	format := fs.String("format", "json", "output format: json | short")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}
	r := roller.NewSeeded(s)
	sys, err := stars.GenerateSystem(r, stars.GenerateSystemOpts{
		WithVariance: true,
		Accuracy:     2,
	})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	switch *format {
	case "json":
		form := stars.BuildSurveyForm(sys, stars.SurveyMetadata{})
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(form)
	case "short":
		_, err := fmt.Fprintln(stdout, stars.ShortProfile(sys))
		return err
	default:
		return fmt.Errorf("unknown format: %q (want json or short)", *format)
	}
}
