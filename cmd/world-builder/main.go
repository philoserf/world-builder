// Command world-builder generates a Mongoose Traveller World Builder's Handbook
// star system from a seed and renders it as Markdown (default), JSON,
// or a short profile string.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/philoserf/world-builder/iiss"
	"github.com/philoserf/world-builder/stars"
	"github.com/philoserf/world-builder/worlds"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("world-builder", flag.ContinueOnError)
	fs.SetOutput(stderr)
	seed := fs.Int64("seed", 0, "random seed (omit for time-based)")
	format := fs.String("format", "markdown", "output format: markdown | json | short")
	peculiar := fs.String("peculiar", "special", "column for Special (2D=2) primary rolls: special | unusual | peculiar")
	if err := fs.Parse(args); err != nil {
		return err
	}

	column, err := stars.ParsePeculiarPath(*peculiar)
	if err != nil {
		return err
	}

	// Distinguish "flag omitted" from "-seed 0": an explicit 0 is a
	// legitimate reproducible seed, not a request for time-based.
	seedSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seedSet = true
		}
	})
	s := *seed
	if !seedSet {
		s = time.Now().UnixNano()
	}

	u, err := worlds.GenerateWithOpts(s, worlds.GenerateOpts{PeculiarColumn: column})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	switch *format {
	case "markdown":
		_, err := fmt.Fprint(stdout, iiss.MarkdownSystem(u.Detail.SystemForms))
		return err
	case "json":
		// Emit the full SystemForms aggregate (Class0I + Class23 + Class4P
		// plus profile strings and mainworld designation) so downstream
		// tooling has everything in one document. Per docs/next-steps.md
		// § B3.
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(u.Detail.SystemForms); err != nil {
			return fmt.Errorf("json: %w", err)
		}
		return nil
	case "short":
		_, err := fmt.Fprintln(stdout, u.Detail.ShortProfile)
		return err
	default:
		return fmt.Errorf("unknown format: %q (want markdown, json, or short)", *format)
	}
}
