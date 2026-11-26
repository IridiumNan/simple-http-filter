package rule

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "embed"
)

func handleInfo(ru *Rule) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "config until: %s\n%s", time.Now().String(), ru.Str())
	}
}

func handleAcceptReload(ru *Rule) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		ipStr := vals.Get("ip")
		if ipStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("ip address should not be empty"))
			return
		}
		err := ru.AddAccept(ipStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, "update successfully, new config: \n%s", ru.Str())
	}
}

func handleDropReload(ru *Rule) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		ipStr := vals.Get("ip")

		if ipStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("ip address should not be empty"))
			return
		}
		err := ru.AddDrop(ipStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, "update successfully, new config: \n%s", ru.Str())
	}
}

var confGen = 0

func handleSave(ru *Rule) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newGenName := fmt.Sprintf("config-%d.json", confGen)
		err := os.Rename("config.json", newGenName)
		if os.IsNotExist(err) {
			slog.Info("old configuration not found, fail to backup")
		} else {
			slog.Info("backup old configuration", "path", newGenName)
			confGen++
		}

		data := ru.Byte()
		if data == nil {
			slog.Error("while marshal rule, skip saving")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "failed to convert rule into byte, fail saving")
			return
		}
		err = os.WriteFile("config.json", data, 0o644)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "while save current configuration into config.json, err: %s", err)
			slog.Error("while save current configuration into config.json", "err", err)

			err := os.Rename(newGenName, "config.json")
			if err != nil {
				slog.Error("failed to recover old config", "err", err)
				fmt.Fprintf(w, "failed to recover old config, err: %s", err)
			}
			return
		}

		fmt.Fprintf(w, "save config success, backup old generation config as: %s", newGenName)
	}
}

func handleReload(ch chan<- Sig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "receive message to reload, try to reload now")

		ch <- SigReload
	}
}

func handleShutdown(ch chan<- Sig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "receive message to shutdown, try to shutdown now")

		ch <- SigShutdown
	}
}

//go:embed API.md
var textDoc []byte

//go:embed API.html
var htmlDoc []byte

func handleAPIDoc(format string) http.HandlerFunc {
	switch format {
	case "html":
		return func(w http.ResponseWriter, r *http.Request) {
			w.Write(htmlDoc)
		}
	default:
		return func(w http.ResponseWriter, r *http.Request) {
			w.Write(textDoc)
		}
	}
}

type Sig uint8

const (
	SigReload Sig = iota
	SigShutdown
)

// NewConfigServe
// return a handler with
// /info get informations about current config
//
// /modify/accept?ip=192.168.1.1
// /modify/drop?ip=192.168.1.1
func NewConfigServe(r *Rule, ch chan<- Sig) *http.Server {
	h := http.NewServeMux()

	h.HandleFunc("/info", handleInfo(r))
	h.HandleFunc("/modify/accept", handleAcceptReload(r))
	h.HandleFunc("/modify/drop", handleDropReload(r))

	h.HandleFunc("/api/text", handleAPIDoc("text"))
	h.HandleFunc("/api/html", handleAPIDoc("html"))

	h.HandleFunc("/save", handleSave(r))
	h.HandleFunc("/reload", handleReload(ch))
	h.HandleFunc("/shutdown", handleShutdown(ch))

	return &http.Server{
		Handler: h,
	}
}
