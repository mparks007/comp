package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/mparks007/comp/csvfilepractice/states"
)

const stateInfoHtmlFile = "stateinfo.html"

func main() {
	// have to at least have a source state file
	if len(os.Args) < 2 {
		log.Fatalln("Missing source file parameter.")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer f.Close()

	// read all the states info
	var statesInfo states.StatesFileInfo
	parseErrors, err := statesInfo.ReadAndParse(csv.NewReader(f))
	if err != nil {
		log.Fatalln("Error reading from file:", err)
	}

	// write out any line-specific parse issues
	if parseErrors != nil {
		fmt.Println("The following specific parse errors occurred:")
		for _, parseError := range parseErrors {
			log.Println(parseError)
		}
		fmt.Println()
	}

	// write states info to console
	fmt.Print(statesInfo.String())

	// optionally, lookup an abbreviation
	if len(os.Args) == 3 {
		if abbrev, ok := statesInfo.LookupState(os.Args[2]); ok {
			fmt.Printf("\nState \"%s\" is \"%s\"\n", os.Args[2], abbrev)
		} else {
			fmt.Printf("\nState \"%s\" not found in state table: %s\n", os.Args[2], os.Args[1])
		}
	}

	f2, err := os.Create(stateInfoHtmlFile)
	if err != nil {
		log.Fatalln("Error opening %s: %s\n", stateInfoHtmlFile, err.Error())
		return
	}
	defer f2.Close()

	// dump the results to an html page
	statesInfo.WriteHtmlPage(f2)
}
