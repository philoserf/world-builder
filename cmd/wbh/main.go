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
	"wbh/worlds"
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
	format := fs.String("format", "markdown", "output format: markdown | json | short")
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
	case "markdown":
		sp, err := worlds.GenerateSystemPlacement(r, sys)
		if err != nil {
			return fmt.Errorf("system placement: %w", err)
		}
		sd, err := worlds.DetailSystem(r, sys, sp, worlds.IISSClass23Header{})
		if err != nil {
			return fmt.Errorf("detail system: %w", err)
		}
		_, err = fmt.Fprint(stdout, worlds.RenderSystemMarkdown(sd, sys))
		return err
	case "json":
		form := stars.BuildSurveyForm(sys, stars.SurveyMetadata{})
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(form)
	case "short":
		_, err := fmt.Fprintln(stdout, stars.ShortProfile(sys))
		return err
	default:
		return fmt.Errorf("unknown format: %q (want markdown, json, or short)", *format)
	}
}
