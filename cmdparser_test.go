package main

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	tests := map[string]struct {
		args    []string
		wantErr bool
		want    *parsed
	}{
		"show version short": {
			args:    []string{"", "-v"},
			wantErr: false,
			want: &parsed{
				tags:          []string{},
				note:          "",
				showVersion:   true,
				showHelp:      false,
				readRequested: false,
			},
		},
		"show version long": {
			args:    []string{"", "-version"},
			wantErr: false,
			want: &parsed{
				tags:          []string{},
				note:          "",
				showVersion:   true,
				showHelp:      false,
				readRequested: false,
			},
		},
		"show help short": {
			args:    []string{"", "-h"},
			wantErr: false,
			want: &parsed{
				tags:          []string{},
				note:          "",
				showVersion:   false,
				showHelp:      true,
				readRequested: false,
			},
		},
		"combine short": {
			args:    []string{"", "-h", "-v"},
			wantErr: false,
			want: &parsed{
				tags:          []string{},
				note:          "",
				showVersion:   true,
				showHelp:      true,
				readRequested: false,
			},
		},
		"show help long": {
			args:    []string{"", "-help"},
			wantErr: false,
			want: &parsed{
				tags:          []string{},
				note:          "",
				showVersion:   false,
				showHelp:      true,
				readRequested: false,
			},
		},
		"no flags": {
			args:    []string{""},
			wantErr: false,
			want: &parsed{
				tags:          []string{},
				note:          "",
				showVersion:   false,
				showHelp:      false,
				readRequested: true,
			},
		},
		"erroneous flag": {
			args:    []string{"", "-this-will-not-be-valid-flag"},
			wantErr: true,
			want:    nil,
		},
	}

	t.Parallel()

	for name, tc := range tests {
		testCase := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			parser := newCmdParser("", io.Discard)
			parsed, err := parser.parse(testCase.args)
			if testCase.wantErr && err == nil {
				t.Fatal("expected parse error but got nil")
			}
			if !reflect.DeepEqual(testCase.want, parsed) {
				t.Fatalf("\n\texpected parsed:\n\t\t%#v\n\tgot:\n\t\t%#v",
					testCase.want,
					parsed)
			}
		})
	}
}

func TestShowUsage(t *testing.T) {
	t.Parallel()

	var buf strings.Builder

	const name = "1337"

	parser := newCmdParser(name, &buf)
	parser.showUsage()

	if got := buf.String(); !strings.Contains(got, "Usage") || !strings.Contains(got, name) {
		t.Fatalf("wanted usage to contain [Usage, %v] but got %#v", name, got)
	}
}

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
		"multiple args multiple tags": {
			input:        []string{"@x ", "y", "z ", " @w"},
			expectedTags: []string{"w", "x"},
			expectedNote: "y z",
		},
	}

	t.Parallel()

	for name, tc := range tests {
		testCase := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tags, note := parseArgs(testCase.input)
			if testCase.expectedNote != note {
				t.Fatalf("\n\texpected note:\n\t\t%#v\n\tgot:\n\t\t%#v",
					testCase.expectedNote,
					note)
			}
			if !reflect.DeepEqual(testCase.expectedTags, tags) {
				t.Fatalf("\n\texpected tags:\n\t\t%#v\n\tgot:\n\t\t%#v",
					testCase.expectedTags,
					tags)
			}
		})
	}
}
