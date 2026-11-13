package display

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"time"
)

const (
	welcomeBannerLen = 8
	welcomeSpinSteps = 100
	welcomeSpinDelay = 30 * time.Millisecond
)

// ShowWelcome animates random characters on dev until a non-loopback IPv4
// address is available or ctx is cancelled, then displays the IP suffix banner.
func ShowWelcome(ctx context.Context, dev Device) error {
	return showWelcome(ctx, dev, resolveIPv4, welcomeSpinDelay)
}

func showWelcome(
	ctx context.Context,
	dev Device,
	ipResolver func() (string, error),
	stepDelay time.Duration,
) error {
	for {
		for range welcomeSpinSteps {
			dev.SetString(randomBanner())
			if err := sleepContext(ctx, stepDelay); err != nil {
				return err
			}
		}

		ip, err := ipResolver()
		if err != nil || ip == "" {
			continue
		}

		dev.SetString(formatWelcomeBanner(ip))
		return nil
	}
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func formatWelcomeBanner(ip string) string {
	suffix := ip
	if len(suffix) > 3 {
		suffix = suffix[len(suffix)-3:]
	}
	return fmt.Sprintf("flwr-%3s", suffix)
}

func randomBanner() string {
	const (
		minASCII = 32
		maxASCII = 126
	)
	buf := make([]byte, welcomeBannerLen)
	for i := range buf {
		buf[i] = byte(minASCII + rand.IntN(maxASCII-minASCII))
	}
	return string(buf)
}

func resolveIPv4() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("list network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		isDown := (iface.Flags & net.FlagUp) == 0
		isLoopback := (iface.Flags & net.FlagLoopback) != 0
		if isDown || isLoopback {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ip := extractIPv4(addr); ip != nil {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("cannot resolve the IP")
}

func extractIPv4(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil {
		return nil
	}
	return ip.To4()
}
