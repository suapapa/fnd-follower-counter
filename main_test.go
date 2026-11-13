package main

import (
	"io"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    cliOptions
		wantErr bool
	}{
		{
			name: "default flags",
			args: nil,
			want: cliOptions{
				configPath: "config.yaml",
				console:    false,
			},
		},
		{
			name: "custom config flag",
			args: []string{"-c", "my.yaml"},
			want: cliOptions{
				configPath: "my.yaml",
				console:    false,
			},
		},
		{
			name: "console flag",
			args: []string{"-n"},
			want: cliOptions{
				configPath: "config.yaml",
				console:    true,
			},
		},
		{
			name: "custom config and console flags",
			args: []string{"-c", "my.yaml", "-n"},
			want: cliOptions{
				configPath: "my.yaml",
				console:    true,
			},
		},
		{
			name:    "positional config path rejected",
			args:    []string{"my.yaml"},
			wantErr: true,
		},
		{
			name:    "unknown flag rejected",
			args:    []string{"-t", "50"},
			wantErr: true,
		},
		{
			name:    "extra positional args rejected",
			args:    []string{"-c", "a.yaml", "b.yaml"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFlags(tc.args, io.Discard)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseFlags(%v) err = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Fatalf("parseFlags(%v) = %+v, want %+v", tc.args, got, tc.want)
			}
		})
	}
}
