package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"time"

	_ "embed"

	"simple-http/filter/internal/proxy"
	"simple-http/filter/internal/rule"
)

var LogLevel = slog.LevelInfo

//go:embed config.json
var configByte []byte

//go:embed help.txt
var helpByte []byte

var configPath string = "config.json"

func printHelp() {
	syscall.Write(1, helpByte)
}

func generateConf() {
	err := os.WriteFile("config.json", configByte, 0o644)
	if err != nil {
		slog.Error("error while creating config file", "err", err)
		os.Exit(1)
	}
	fmt.Printf("generate default configuration file on path -> %s", "config.json")
}

func startProxy(ru *rule.Rule, ch <-chan rule.Sig) {
	p := proxy.NewProxy(ru)

	addr := fmt.Sprintf("%s:%s", ru.ServeHost, ru.ServePort)

	slog.Info("starting serve")
	slog.Info(ru.Str())

	proxyServe := http.Server{
		Addr:    addr,
		Handler: p,
	}

	go func() {
		err := proxyServe.ListenAndServe()
		if err == http.ErrServerClosed {
			slog.Info("proxy server graceful closed")
		} else {
			slog.Error("while proxy server listen and serve", "err", err)
		}
	}()

	go func() {
		<-ch
		slog.Info("receive shutdown request, try to shutdown proxy server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		proxyServe.Shutdown(ctx)
	}()
}

func startConfServer(ru *rule.Rule, controlCh chan<- rule.Sig, shutCh <-chan rule.Sig) {
	confServe := rule.NewConfigServe(ru, controlCh)
	confServe.Addr = ru.ConfigServeAddr

	slog.Info("starting reload serve", "addr", ru.ConfigServeAddr)
	go func() {
		err := confServe.ListenAndServe()
		if err != nil {
			slog.Warn("error while starting reload serve", "err", err, "addr", ru.ConfigServeAddr)
		}
	}()

	go func() {
		<-shutCh
		slog.Info("receive reload request, try to reload the config server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		confServe.Shutdown(ctx)
	}()
}

func main() {
	slog.SetLogLoggerLevel(LogLevel)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			generateConf()
			return
		case "help", "-h", "--help":
			printHelp()
			return
		default:
			configPath = os.Args[1]
		}
	}

	// TODO: Write an endpoint for graceful reload
	// /reload
	// read configuration from file

	for {
		controlCh := make(chan rule.Sig)
		shutCh := make(chan rule.Sig)

		f, err := os.OpenFile(configPath, os.O_RDONLY, 0o644)
		if err != nil {
			slog.Error("error while opening configuration file", "err", err, "path", configPath)
			os.Exit(1)
		}

		defer f.Close()
		ru, err := rule.NewRule(f)
		if err != nil {
			log.Fatalf("error while reading rule config: %s", err)
		}
		startProxy(ru, shutCh)
		startConfServer(ru, controlCh, shutCh)

		act := <-controlCh

		switch act {
		case rule.SigReload:
			close(shutCh)
		case rule.SigShutdown:
			return
		}

	}
}
