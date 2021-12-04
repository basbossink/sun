package cmdparser

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := map[string]struct {
		input        []string
		expectedTags []string
		expectedNote string
	}{
		"single space": {
			input:        []string{" "},
			expectedTags: []string{},
			expectedNote: "",
		},
		"single word": {
			input:        []string{"x"},
			expectedTags: []string{},
			expectedNote: "x",
		},
		"single arg multiple words": {
			input:        []string{"x x"},
			expectedTags: []string{},
			expectedNote: "x x",
		},
		"multiple args single word": {
			input:        []string{"x", " "},
			expectedTags: []string{},
			expectedNote: "x",
		},
		"multiple args multiple words": {
			input:        []string{"x", "y"},
			expectedTags: []string{},
			expectedNote: "x y",
		},
		"multiple args multiple words should trim": {
			input:        []string{"x ", " y"},
			expectedTags: []string{},
			expectedNote: "x y",
		},
		"single tag": {
			input:        []string{"@x"},
			expectedTags: []string{"x"},
			expectedNote: "",
		},
		"multiple args single tag": {
			input:        []string{"@x ", "y"},
			expectedTags: []string{"x"},
			expectedNote: "y",
		},
		"single arg multiple words single tag": {
			input:        []string{"x @y x"},
			expectedTags: []string{"y"},
			expectedNote: "x x",
		},
		"multiple args mulitple tags": {
			input:        []string{"@x ", "y", "z ", " @w"},
			expectedTags: []string{"w", "x"},
			expectedNote: "y z",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tags, note := parseArgs(tc.input)
			if tc.expectedNote != note {
				t.Fatalf("\n\texpected note:\n\t\t%#v\n\tgot:\n\t\t%#v",
					tc.expectedNote,
					note)
			}
			if !reflect.DeepEqual(tc.expectedTags, tags) {
				t.Fatalf("\n\texpected tags:\n\t\t%#v\n\tgot:\n\t\t%#v",
					tc.expectedTags,
					tags)
			}
		})
	}
}
