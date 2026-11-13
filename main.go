package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/suapapa/fnd-follower-counter/internal/config"
	"github.com/suapapa/fnd-follower-counter/internal/display"
	"github.com/suapapa/fnd-follower-counter/internal/sns"
	"github.com/suapapa/go_devices/tm1638"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

const (
	defaultConfigPath = "config.yaml"
	pollTimeout       = 25 * time.Second
	pinData           = "17"
	pinClock          = "27"
	pinStrobe         = "22"
)

type cliOptions struct {
	configPath string
	console    bool
}

type sideRunner struct {
	label   string
	target  config.Target
	fetcher sns.Fetcher
	state   display.SideState
}

func newSideRunner(label string, side config.SideConfig) (*sideRunner, error) {
	target, err := side.Resolve()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	fetcher, err := sns.NewFetcher(target.Provider)
	if err != nil {
		return nil, fmt.Errorf("%s (%s): %w", label, target.Provider, err)
	}
	return &sideRunner{
		label:   label,
		target:  target,
		fetcher: fetcher,
	}, nil
}

func (r *sideRunner) poll(ctx context.Context) {
	cnt, err := r.fetcher.FetchFollowerCount(ctx, r.target.Handle)
	if err != nil {
		r.state.Err = err
		log.Printf("[%s] error: %v", r.logTag(), err)
		return
	}
	r.state = display.SideState{
		Count:    cnt,
		HasValue: true,
	}
	log.Printf("[%s] count=%d", r.logTag(), cnt)
}

func (r *sideRunner) logTag() string {
	return fmt.Sprintf("%s %s:%s", r.label, r.target.Provider, r.target.Handle)
}

func main() {
	opts, err := parseFlags(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(rootCtx, opts); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func run(ctx context.Context, opts cliOptions) error {
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return err
	}

	left, err := newSideRunner("left", cfg.Left)
	if err != nil {
		return err
	}
	right, err := newSideRunner("right", cfg.Right)
	if err != nil {
		return err
	}

	dev, err := openDisplay(ctx, opts.console)
	if err != nil {
		return err
	}

	refreshSides(
		ctx,
		dev,
		left,
		right,
	)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			refreshSides(
				ctx,
				dev,
				left,
				right,
			)
		case <-ctx.Done():
			log.Println("byebye")
			return nil
		}
	}
}

func openDisplay(ctx context.Context, console bool) (display.Device, error) {
	if console {
		return display.NewConsoleFND(os.Stdout), nil
	}

	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("init host gpio: %w", err)
	}

	dev, err := tm1638.Open(
		gpioreg.ByName(pinData),
		gpioreg.ByName(pinClock),
		gpioreg.ByName(pinStrobe),
	)
	if err != nil {
		return nil, fmt.Errorf("open tm1638: %w", err)
	}
	if err := display.ShowWelcome(ctx, dev); err != nil {
		return nil, err
	}
	return dev, nil
}

func refreshSides(
	ctx context.Context,
	dev display.Device,
	left *sideRunner,
	right *sideRunner,
) {
	pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Go(func() {
		left.poll(pollCtx)
	})
	wg.Go(func() {
		right.poll(pollCtx)
	})
	wg.Wait()

	display.Render(dev, left.state, right.state)
}

func parseFlags(args []string, errOut io.Writer) (cliOptions, error) {
	opts := cliOptions{
		configPath: defaultConfigPath,
	}

	fs := flag.NewFlagSet("fnd-follower-counter", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.StringVar(&opts.configPath, "c", opts.configPath, "path to config YAML file")
	fs.BoolVar(
		&opts.console,
		"n",
		false,
		"render FND display as ASCII art on console instead of TM1638",
	)

	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}
	if fs.NArg() > 0 {
		return cliOptions{}, fmt.Errorf("unexpected positional arguments %v; use -c <config>", fs.Args())
	}
	return opts, nil
}
