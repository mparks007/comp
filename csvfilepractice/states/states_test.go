package states

import "testing"

// state
// -----
// String
//		string returned

// StatesFileInfo
// --------------
// ReadAndParse(mock Reader) [Read method (return mock csv rows, return EOF, return other error)]
//		major error
//		states slice fully loaded / no minor parse erros
//		states slice partially loaded / some minor parse errors
//
//		parseState([]string)
//			state not returned, error
//			state returned, no error
//
// LookupState(string)
//		found, true
//		not found, false
//
// AsHtmlPage
//		if states, page text returned
//		if no states, "" returned
//
// String
//		if states, string returned
//		if no states, "" returned


func Test(t *testing.T) {

}
