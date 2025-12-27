package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitError   = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintf(stderr, "Usage: json-subset <subset.json> <superset.json>\n")
		fmt.Fprintf(stderr, "\nCheck if the first JSON is a subset of the second JSON.\n")
		fmt.Fprintf(stderr, "Arrays are compared as sets (order is ignored).\n")
		return exitError
	}

	subsetFile := args[0]
	supersetFile := args[1]

	subsetData, err := loadJSON(subsetFile)
	if err != nil {
		fmt.Fprintf(stderr, "Error loading %s: %v\n", subsetFile, err)
		return exitError
	}

	supersetData, err := loadJSON(supersetFile)
	if err != nil {
		fmt.Fprintf(stderr, "Error loading %s: %v\n", supersetFile, err)
		return exitError
	}

	isSubset, diffs := checkSubsetWithDiffs(subsetData, supersetData)

	if isSubset {
		fmt.Fprintln(stdout, "OK: First JSON is a subset of second JSON.")
		return exitSuccess
	}

	fmt.Fprintln(stderr, "FAIL: First JSON is not a subset of second JSON.")
	fmt.Fprintln(stderr, "")
	diffOutput := FormatDiffOutput(subsetData, diffs)
	fmt.Fprint(stderr, diffOutput)
	return exitFailure
}

func loadJSON(filename string) (interface{}, error) {
	var data []byte
	var err error

	if filename == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
