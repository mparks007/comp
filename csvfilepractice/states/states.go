package states

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// MyCsvReader is used to allow mocking of "encoding/csv.Reader"
type MyCsvReader interface {
	Read() (record []string, err error)
}

// basic struct to store a subset of state data
type state struct {
	id               int
	name             string
	abbreviation     string
	censusRegionName string
}

// String returns a basic dump of the state data
func (s state) String() string {
	return fmt.Sprintf("ID:%d\tName:%-20s\tAbbreviation:%s\t\tCensus Region:%s\n", s.id, s.name, s.abbreviation, s.censusRegionName)
}

// StatesFileInfo wraps both a column-name-to-index map and states data into one handy object
type StatesFileInfo struct {
	columnXRef map[string]int
	states     []state
}

// ReadAndParse reads the csv file line by line to build up the header line map on first line then states data on the remaining lines
func (s *StatesFileInfo) ReadAndParse(rdr MyCsvReader) (minorErrors []string, majorError error) {
	s.columnXRef = make(map[string]int)
	var parseErrors []string

	for lineCount := 0; ; lineCount++ {
		// read a row at a time from the csv (since care special about the first row)
		fields, err := rdr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		} else if len(fields) == 0 { // sanity check (maybe isn't possible without an error being set on Read)
			return nil, fmt.Errorf("No fields read from file line: %d", lineCount+1)
		}

		// if on first line of file (header line)
		if lineCount == 0 {
			// build a map that tracks which index in the csv fields corresponds to a column name
			for i, columnName := range fields {
				s.columnXRef[columnName] = i
			}
		} else { // non-header line

			// parse out some state data
			st, err := s.parseState(fields)
			if err != nil {
				parseErrors = append(parseErrors, fmt.Sprintf("Data:%v, Error:[%s]", fields, err.Error()))
			} else {
				s.states = append(s.states, st)
			}
		}
	}

	return parseErrors, nil
}

// parseState takes a slice of state data and returns a state struct filled with it
func (s *StatesFileInfo) parseState(fields []string) (state, error) {
	id, err := strconv.Atoi(fields[s.columnXRef["id"]])
	if err != nil {
		return state{}, err
	}

	name := fields[s.columnXRef["name"]]
	abbreviation := fields[s.columnXRef["abbreviation"]]
	censusRegionName := fields[s.columnXRef["census_region_name"]]

	return state{
		id:               id,
		name:             name,
		abbreviation:     abbreviation,
		censusRegionName: censusRegionName,
	}, nil
}

// LookupState returns the state name associated with the passed in abbreviation value, or "" if fails lookup
func (s *StatesFileInfo) LookupState(abbreviation string) (string, bool) {

	for _, state := range s.states {
		if state.abbreviation == abbreviation {
			return state.name, true
		}
	}

	return "", false
}

// AsHtmlPage generates a basic table of the state data
func (s *StatesFileInfo) AsHtmlPage(cssFile string) string {

	// early exit
	if len(s.states) == 0 {
		return ""
	}

	var buffer bytes.Buffer

	// starting html
	buffer.WriteString(fmt.Sprint(`
<html>
	<head>`))

	buffer.WriteString(fmt.Sprintf(`
		<link rel="stylesheet" href="%s">`, cssFile))

	buffer.WriteString(fmt.Sprint(`
	</head>
	<body>
		<table>
			<tr>
				<th>ID</th>
				<th>Abbreviation</th>
				<th>Name</th>
				<th>Census Region</th>
			</tr>`))

	// specific state content
	for _, state := range s.states {
		buffer.WriteString(fmt.Sprintf(`
			<tr>
				<td>%d</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
			</tr>`, state.id, state.abbreviation, state.name, state.censusRegionName))
	}

	// ending html
	buffer.WriteString(fmt.Sprint(`
		</table>
	</body>
</html>`))

	return buffer.String()
}

// Strings writes out a simple, tabbified table of state data
func (s *StatesFileInfo) String() string {

	// early exit
	if len(s.states) == 0 {
		return ""
	}

	var buffer bytes.Buffer

	for _, st := range s.states {
		buffer.WriteString(st.String())
	}

	return buffer.String()
}
