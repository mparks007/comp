package states

import "testing"

// StatesFileInfo
// --------------
// ReadAndParse(mock Reader) [Read method (return mock csv rows, return EOF, return other error)]
//		nothing to parse (immediate EOF)
//		major error
//		states slice fully loaded / no minor parse erros
//		states slice partially loaded / some minor parse errors

func TestStateEmptyAsString(t *testing.T) {

	st := state{}

	want := "ID:0\tName:                    \tAbbreviation:	\tCensus Region:\n"
	if got := st.String(); got != want {
		t.Errorf("Wanted '%s' but got '%s'\n", want, got)
	}
}

func TestStateFilledAsString(t *testing.T) {

	st := state{
		id:               100,
		name:             "MyState",
		abbreviation:     "MyAbbrev",
		censusRegionName: "MyCensus",
	}

	want := "ID:100\tName:MyState             \tAbbreviation:MyAbbrev	\tCensus Region:MyCensus\n"
	if got := st.String(); got != want {
		t.Errorf("Wanted '%s' but got '%s'\n", want, got)
	}
}

type mockReader struct {
	read func (p []byte) (n int, err error)
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	return r.read(p)
}

func TestStatesFileInfoReadAndParseNothingToParse(t *testing.T) {

	info := StatesFileInfo{}

	reader := mockReader{}

	parseErrors, err := info.ReadAndParse(reader)
}

func TestStatesFileInfoReadAndParseMajorError(t *testing.T) {
}

func TestStatesFileInfoReadAndParseTotalSuccess(t *testing.T) {
}

func TestStatesFileInfoReadAndParsePartialSuccess(t *testing.T) {
}

func TestStatesFileInfoLookupStateNotFound(t *testing.T) {

	info := StatesFileInfo{}

	info.states = append(info.states, state{
		id:               100,
		name:             "MyState",
		abbreviation:     "MyAbbrev",
		censusRegionName: "MyCensus",
	})

	wantOk := false
	wantState := ""
	if gotState, gotOk := info.LookupState("XYZ"); gotOk == wantOk {
		if gotState != wantState {
			t.Errorf("Wanted state '%s' but got '%s'\n", wantState, gotState)
		}
	} else {
		t.Errorf("Wanted abbreviation lookup result %t but got %t\n", wantOk, gotOk)
	}
}

func TestStatesFileInfoLookupStateFound(t *testing.T) {

	info := StatesFileInfo{}

	info.states = append(info.states, state{
		id:               100,
		name:             "MyState",
		abbreviation:     "MyAbbrev",
		censusRegionName: "MyCensus",
	})

	info.states = append(info.states, state{
		id:               102,
		name:             "MyState2",
		abbreviation:     "MyAbbrev2",
		censusRegionName: "MyCensus2",
	})

	wantOk := true
	wantState := "MyState2"
	if gotState, gotOk := info.LookupState("MyAbbrev2"); gotOk == wantOk {
		if gotState != wantState {
			t.Errorf("Wanted state '%s' but got '%s'\n", wantState, gotState)
		}
	} else {
		t.Errorf("Wanted abbreviation lookup success %t but got %t\n", wantOk, gotOk)
	}
}

func TestStatesFileInfoAsHtmlPageNoStates(t *testing.T) {
	info := StatesFileInfo{}

	wantHtml := ""
	if gotHtml := info.AsHtmlPage("file.css"); gotHtml != wantHtml {
		t.Errorf("Wanted HTML of '%s' but got '%s'\n", wantHtml, gotHtml)
	}
}

func TestStatesFileInfoAsHtmlPageWithStates(t *testing.T) {
	info := StatesFileInfo{}

	info.states = append(info.states, state{
		id:               100,
		name:             "MyState",
		abbreviation:     "MyAbbrev",
		censusRegionName: "MyCensus",
	})

	info.states = append(info.states, state{
		id:               102,
		name:             "MyState2",
		abbreviation:     "MyAbbrev2",
		censusRegionName: "MyCensus2",
	})

	wantHtml := `
<html>
	<head>
		<link rel="stylesheet" href="file.css">
	</head>
	<body>
		<table>
			<tr>
				<th>ID</th>
				<th>Abbreviation</th>
				<th>Name</th>
				<th>Census Region</th>
			</tr>
			<tr>
				<td>100</td>
				<td>MyAbbrev</td>
				<td>MyState</td>
				<td>MyCensus</td>
			</tr>
			<tr>
				<td>102</td>
				<td>MyAbbrev2</td>
				<td>MyState2</td>
				<td>MyCensus2</td>
			</tr>
		</table>
	</body>
</html>`
	if gotHtml := info.AsHtmlPage("file.css"); gotHtml != wantHtml {
		t.Errorf("Wanted HTML of '%s' but got '%s'\n", wantHtml, gotHtml)
	}
}

func TestStatesFileInfoEmptyAsString(t *testing.T) {

	info := StatesFileInfo{}

	want := ""
	if got := info.String(); got != want {
		t.Errorf("Wanted '%s' but got '%s'\n", want, got)
	}
}

func TestStatesFileInfoFilledAsString(t *testing.T) {

	info := StatesFileInfo{}

	info.states = append(info.states, state{
		id:               100,
		name:             "MyState",
		abbreviation:     "MyAbbrev",
		censusRegionName: "MyCensus",
	})

	info.states = append(info.states, state{
		id:               102,
		name:             "MyState2",
		abbreviation:     "MyAbbrev2",
		censusRegionName: "MyCensus2",
	})

	want := "ID:100\tName:MyState             \tAbbreviation:MyAbbrev	\tCensus Region:MyCensus\nID:102\tName:MyState2            \tAbbreviation:MyAbbrev2	\tCensus Region:MyCensus2\n"
	if got := info.String(); got != want {
		t.Errorf("Wanted '%s' but got '%s'\n", want, got)
	}
}
