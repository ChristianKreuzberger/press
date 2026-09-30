package main

import (
	"flag"
	"io"
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantFile string
		wantPos  []string
		wantErr  bool
	}{
		{"positional first", []string{"foo", "--file", "x.md"}, "x.md", []string{"foo"}, false},
		{"flag first", []string{"--file", "x.md", "foo"}, "x.md", []string{"foo"}, false},
		{"equals form", []string{"foo", "--file=x.md"}, "x.md", []string{"foo"}, false},
		{"two positionals", []string{"a", "--file", "x.md", "b"}, "x.md", []string{"a", "b"}, false},
		{"double dash", []string{"--file", "x.md", "--", "--weird"}, "x.md", []string{"--weird"}, false},
		{"none", nil, "", nil, false},
		{"unknown flag", []string{"foo", "--bogus"}, "", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("t", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			file := fs.String("file", "", "")
			pos, err := parseArgs(fs, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if *file != tt.wantFile {
				t.Errorf("file = %q, want %q", *file, tt.wantFile)
			}
			if len(pos) != len(tt.wantPos) || (len(pos) > 0 && !reflect.DeepEqual(pos, tt.wantPos)) {
				t.Errorf("positionals = %v, want %v", pos, tt.wantPos)
			}
		})
	}
}
