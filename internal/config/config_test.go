package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		wantInterval time.Duration
		wantLeft     Target
		wantRight    Target
		wantErr      bool
	}{
		{
			name: "youtube and insta default interval",
			yaml: `
left:
  youtube: "@vibetinker"
right:
  insta: homin.crc
`,
			wantInterval: 60 * time.Second,
			wantLeft:     Target{Provider: "youtube", Handle: "@vibetinker"},
			wantRight:    Target{Provider: "insta", Handle: "homin.crc"},
		},
		{
			name: "instagram alias and twitch with custom interval",
			yaml: `
interval: 30s
left:
  instagram: homin.crc
right:
  twitch: suapapa
`,
			wantInterval: 30 * time.Second,
			wantLeft:     Target{Provider: "instagram", Handle: "homin.crc"},
			wantRight:    Target{Provider: "twitch", Handle: "suapapa"},
		},
		{
			name: "missing right side",
			yaml: `
left:
  youtube: "@vibetinker"
`,
			wantErr: true,
		},
		{
			name: "multiple providers on one side",
			yaml: `
left:
  youtube: "@vibetinker"
  twitch: suapapa
right:
  insta: homin.crc
`,
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			yaml:    "left: [invalid",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tc.yaml), 0o600); err != nil {
				t.Fatalf("write temp config: %v", err)
			}

			cfg, err := Load(path)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Load() err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			if cfg.Interval != tc.wantInterval {
				t.Fatalf("Interval = %v, want %v", cfg.Interval, tc.wantInterval)
			}
			left, err := cfg.Left.Resolve()
			if err != nil || left != tc.wantLeft {
				t.Fatalf("Left = %+v (err=%v), want %+v", left, err, tc.wantLeft)
			}
			right, err := cfg.Right.Resolve()
			if err != nil || right != tc.wantRight {
				t.Fatalf("Right = %+v (err=%v), want %+v", right, err, tc.wantRight)
			}
		})
	}
}
