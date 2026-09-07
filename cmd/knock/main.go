package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ph-py/knock-go/internal/adapters/builder"
	"github.com/ph-py/knock-go/internal/bootstrap"
	"github.com/ph-py/knock-go/internal/domain"
	"github.com/ph-py/knock-go/internal/service"
)

var version = "1.0.0"

func printUsage() {
	fmt.Println("usage: knock [options] <host> <port[:proto]> [port[:proto]] ...")
	fmt.Println("options:")
	fmt.Println("  -u, --udp            make all ports hits use UDP (default is TCP)")
	fmt.Println("  -d, --delay <t>      wait <t> milliseconds between port hits")
	fmt.Println("  -4, --ipv4           Force usage of IPv4")
	fmt.Println("  -6, --ipv6           Force usage of IPv6")
	fmt.Println("  -v, --verbose        be verbose")
	fmt.Println("  -V, --version        display version")
	fmt.Println("  -h, --help           this help")
	fmt.Println()
	fmt.Println("example:  knock myserver.example.com 123:tcp 456:udp 789:tcp")
	fmt.Println()
	os.Exit(1)
}

func printVersion() {
	fmt.Printf("knock %s\n", version)
	fmt.Println("Copyright (C) 2004-2012 Judd Vinet <jvinet@zeroflux.org>")
	os.Exit(0)
}

type cliOptions struct {
	verbose   bool
	udp       bool
	delayMs   int
	ipVersion domain.IPVersion
	host      string
	ports     []string
}

func parseArgs(args []string) (*cliOptions, error) {
	opts := &cliOptions{
		delayMs:   0,
		ipVersion: domain.IPDefault,
		ports:     make([]string, 0),
	}

	var positional []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			printUsage()
		}
		if arg == "-V" || arg == "--version" {
			printVersion()
		}
		if arg == "-v" || arg == "--verbose" {
			opts.verbose = true
			i++
			continue
		}
		if arg == "-u" || arg == "--udp" {
			opts.udp = true
			i++
			continue
		}
		if arg == "-4" || arg == "--ipv4" {
			opts.ipVersion = domain.IPv4Only
			i++
			continue
		}
		if arg == "-6" || arg == "--ipv6" {
			opts.ipVersion = domain.IPv6Only
			i++
			continue
		}
		if arg == "-d" || arg == "--delay" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("option '%s' requires an argument", arg)
			}
			val, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid delay value: %s", args[i+1])
			}
			opts.delayMs = val
			i += 2
			continue
		}
		if strings.HasPrefix(arg, "-d") && len(arg) > 2 {
			val, err := strconv.Atoi(arg[2:])
			if err != nil {
				return nil, fmt.Errorf("invalid delay value: %s", arg[2:])
			}
			opts.delayMs = val
			i++
			continue
		}
		if strings.HasPrefix(arg, "--delay=") {
			valStr := strings.TrimPrefix(arg, "--delay=")
			val, err := strconv.Atoi(valStr)
			if err != nil {
				return nil, fmt.Errorf("invalid delay value: %s", valStr)
			}
			opts.delayMs = val
			i++
			continue
		}

		if strings.HasPrefix(arg, "-") {
			return nil, fmt.Errorf("unknown option: %s", arg)
		}

		// Argumento posicional
		positional = append(positional, arg)
		i++
	}

	if len(positional) < 2 {
		printUsage()
	}

	opts.host = positional[0]
	opts.ports = positional[1:]

	if opts.delayMs < 0 {
		return nil, fmt.Errorf("error: delay cannot be negative")
	}

	return opts, nil
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	defaultProto := domain.ProtocolTCP
	if opts.udp {
		defaultProto = domain.ProtocolUDP
	}

	// Constrói a sequência de batidas usando o Builder Pattern
	seqBuilder := builder.NewKnockSequenceBuilder().
		SetHost(opts.host).
		SetDefaultProtocol(defaultProto).
		SetDelay(time.Duration(opts.delayMs) * time.Millisecond).
		AddTargetsFromStrings(opts.ports)

	seq, err := seqBuilder.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building knock sequence: %s\n", err)
		os.Exit(1)
	}

	// Inicializa a aplicação via Composition Root (Bootstrap)
	app := bootstrap.Bootstrap(bootstrap.Config{
		Verbose: opts.verbose,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := service.ExecuteKnockSequenceCommand{
		ExecutionID: fmt.Sprintf("knock-%d", time.Now().UnixNano()),
		Host:        opts.host,
		Sequence:    *seq,
		IPVersion:   opts.ipVersion,
		Verbose:     opts.verbose,
	}

	if err := app.Execute(ctx, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
