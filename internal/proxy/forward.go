package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"simple-http/filter/internal/rule"
)

type Proxy struct {
	ru *rule.Rule
}

func (pxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !pxy.ru.F(r) {
		slog.Warn("drop invalid traffic", "addr", r.RemoteAddr)

		w.WriteHeader(http.StatusBadGateway)
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		fmt.Fprintf(w, "your ip address is %s, not allowed to visit this website. contact the admin", ip)

		return
	}
	targetURL := fmt.Sprintf("http://%s:%s%s", pxy.ru.DstHost, pxy.ru.DstPort, r.RequestURI)

	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		slog.Error("while creating new request", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error when creating request"))
		return
	}

	req.Header = r.Header.Clone()

	// append the X-Forward-For field
	commingIp, _, _ := strings.Cut(r.RemoteAddr, ":")
	req.Header.Add("X-Forward-For", commingIp)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("while getting response", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error when getting response"))
		return
	}

	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

func NewProxy(r *rule.Rule) *Proxy {
	return &Proxy{ru: r}
}
