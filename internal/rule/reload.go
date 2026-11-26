package rule

import (
	"fmt"
	"net/http"
	"time"
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

	h.HandleFunc("/reload", handleReload(ch))
	h.HandleFunc("/shutdown", handleShutdown(ch))

	return &http.Server{
		Handler: h,
	}
}
