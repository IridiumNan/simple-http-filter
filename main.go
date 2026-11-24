package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"syscall"

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
		log.Fatalf("error while creating config file", "err", err)
	}
	fmt.Printf("generate default configuration file on path -> %s", "config.json")
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
	f, err := os.OpenFile(configPath, os.O_RDONLY, 0o644)
	if err != nil {
		log.Fatalf("error while opening configuration file", "err", err, "path", configPath)
	}

	defer f.Close()
	ru, err := rule.NewRule(f)
	if err != nil {
		log.Fatalf("error while reading rule config: %s", err)
	}

	p := proxy.NewProxy(ru)

	addr := fmt.Sprintf("%s:%s", ru.ServeHost, ru.ServePort)

	slog.Info("starting serve")

	slog.Info(ru.Str())
	log.Fatal(http.ListenAndServe(addr, p))
}
