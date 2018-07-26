package states

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
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

type StatesFileInfo struct {
	columnXRef map[string]int
	states     []state
}

func (s *StatesFileInfo) ReadAndParse(rdr *csv.Reader) (minorErrors []string, majorError error) {
	s.columnXRef = make(map[string]int)
	var parseErrors []string

	for lineCount := 0; ; lineCount++ {
		// read a row at a time from the csv (since care special about the first row)
		fields, err := rdr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
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

func (s *StatesFileInfo) LookupState(abbreviation string) (string, bool) {

	for _, state := range s.states {
		if state.abbreviation == abbreviation {
			return state.name, true
		}
	}

	return "", false
}

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
