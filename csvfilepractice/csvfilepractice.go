package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

type state struct {
	id               int
	name             string
	abbreviation     string
	censusRegionName string
}

func (s state) String() string {
	return fmt.Sprintf("ID:%d\tName:%-20s\tAbbreviation:%s\t\tCensus Region:%s\n", s.id, s.name, s.abbreviation, s.censusRegionName)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Missing source file parameter.")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer f.Close()

	// dump the results to the console
	states := readAndParse(csv.NewReader(f))
	for _, st := range states {
		fmt.Print(st)
	}

	if len(os.Args) == 3 {
		if abbrev, ok := lookupState(os.Args[2], states); ok {
			fmt.Printf("State \"%s\" is \"%s\"\n", os.Args[2], abbrev)
		} else {
			fmt.Printf("State \"%s\" not found in state table: %s\n", os.Args[2],  os.Args[1])
		}
	}
}

func readAndParse(rdr *csv.Reader) []state {
	var states []state
	columnXRef := make(map[string]int)

	for lineCount := 0; ; lineCount++ {
		// read a row at a time from the csv
		fields, err := rdr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln("Error reading from file:", err)
		}

		// if on header line
		if lineCount == 0 {
			// build a map that tracks which index in the csv fields corresponds to a column name
			for i, columnName := range fields {
				columnXRef[columnName] = i
			}
		} else { // non-header line

			// parse out some state data
			st, err := parseState(columnXRef, fields)
			if err != nil {
				log.Println("Error parsing csv row:", err)
			}
			states = append(states, st)
		}
	}
	return states
}

func parseState(columnXRef map[string]int, fields []string) (state, error) {
	id, err := strconv.Atoi(fields[columnXRef["id"]])
	if err != nil {
		return state{}, err
	}

	name := fields[columnXRef["name"]]
	abbreviation := fields[columnXRef["abbreviation"]]
	censusRegionName := fields[columnXRef["census_region_name"]]

	return state{
		id:               id,
		name:             name,
		abbreviation:     abbreviation,
		censusRegionName: censusRegionName,
	}, nil
}

func lookupState(abbreviation string, states []state) (string, bool) {
 
	for _, state := range states {
		if state.abbreviation == abbreviation {
			return state.name, true
		}
	}

	return "", false
}