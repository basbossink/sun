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
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			parser := newCmdParser("", io.Discard)
			parsed, err := parser.parse(tc.args)
			if tc.wantErr && err == nil {
				t.Fatal("expected parse error but got nil")
			}
			if !reflect.DeepEqual(tc.want, parsed) {
				t.Fatalf("\n\texpected parsed:\n\t\t%#v\n\tgot:\n\t\t%#v",
					tc.want,
					parsed)
			}
		})
	}
}

func TestShowUsage(t *testing.T) {
	var buf strings.Builder
	const name = "1337"
	parser := newCmdParser(name, &buf)
	parser.showUsage()
	got := buf.String()
	if !strings.Contains(got, "Usage") || !strings.Contains(got, name) {
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
