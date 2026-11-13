package display

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestFormatWelcomeBanner(t *testing.T) {
	tests := []struct {
		ip   string
		want string
	}{
		{ip: "192.168.1.42", want: "flwr-.42"},
		{ip: "10.0.0.123", want: "flwr-123"},
		{ip: "12", want: "flwr- 12"},
	}

	for _, tc := range tests {
		if got := formatWelcomeBanner(tc.ip); got != tc.want {
			t.Errorf("formatWelcomeBanner(%q) = %q, want %q", tc.ip, got, tc.want)
		}
	}
}

func TestRandomBannerLength(t *testing.T) {
	got := randomBanner()
	if len(got) != welcomeBannerLen {
		t.Fatalf("len(randomBanner()) = %d, want %d", len(got), welcomeBannerLen)
	}
}

func TestShowWelcome(t *testing.T) {
	t.Run("displays IP suffix on success", func(t *testing.T) {
		dev := &mockFND{}
		resolver := func() (string, error) {
			return "192.168.0.105", nil
		}

		err := showWelcome(context.Background(), dev, resolver, time.Microsecond)
		if err != nil {
			t.Fatalf("showWelcome() err = %v", err)
		}
		if len(dev.strings) != welcomeSpinSteps+1 {
			t.Fatalf("len(dev.strings) = %d, want %d", len(dev.strings), welcomeSpinSteps+1)
		}
		gotLast := dev.strings[len(dev.strings)-1]
		if gotLast != "flwr-105" {
			t.Fatalf("last banner = %q, want %q", gotLast, "flwr-105")
		}
	})

	t.Run("stops when context cancelled", func(t *testing.T) {
		dev := &mockFND{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		resolver := func() (string, error) {
			return "", errors.New("no network")
		}

		err := showWelcome(ctx, dev, resolver, time.Millisecond)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("showWelcome() err = %v, want context.Canceled", err)
		}
	})
}

func TestExtractIPv4(t *testing.T) {
	ip4 := &net.IPNet{IP: net.ParseIP("192.168.1.10")}
	if got := extractIPv4(ip4); got == nil || got.String() != "192.168.1.10" {
		t.Fatalf("extractIPv4(v4) = %v, want 192.168.1.10", got)
	}

	ip6 := &net.IPNet{IP: net.ParseIP("fe80::1")}
	if got := extractIPv4(ip6); got != nil {
		t.Fatalf("extractIPv4(v6) = %v, want nil", got)
	}
}
