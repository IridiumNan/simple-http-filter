package rule

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"

	"simple-http/filter/internal/filter"
)

type Rule struct {
	ConfigServeAddr string `json:"config_serve_addr"`

	ServeHost string `json:"serve_host"`
	ServePort string `json:"serve_port"`

	DstHost string `json:"dst_host"`
	DstPort string `json:"dst_port"`

	// Can be accept or drop
	DefaultPolicy string `json:"default_policy"`

	AcceptHosts []string `json:"accept_hosts"`

	DropHosts []string `json:"drop_hosts"`

	acceptMatcher []Matcher
	dropMatcher   []Matcher

	F filter.Filter

	mu sync.RWMutex
}

func (r *Rule) Str() string {
	return fmt.Sprintf(`
		Using config as below

		listen: %s:%s
		forward_dst: http://%s:%s

		default_policy: %s

		accept_hosts: %v
		
		drop_hosts: %v
		`, r.ServeHost, r.ServePort,
		r.DstHost, r.DstPort,
		r.DefaultPolicy,
		r.AcceptHosts,
		r.DropHosts)
}

func (r *Rule) buildFilter() {
	r.mu.Lock()
	switch strings.ToLower(r.DefaultPolicy) {
	case "accept":
		r.buildAcceptFilter()
	case "drop":
		r.buildDropFilter()
	default:
		log.Fatalf("invalid policy!!!, policy must be accept or drop, got: %s", r.DefaultPolicy)
	}
	r.mu.Unlock()
}

func (r *Rule) buildAcceptFilter() {
	r.dropMatcher = make([]Matcher, 0, len(r.DropHosts))
	for _, h := range r.DropHosts {
		m, err := GenMatcher(h)
		if err != nil {
			slog.Error("failed to generate mathcer from host, skip", "raw", h, "err", err)
		}
		r.dropMatcher = append(r.dropMatcher, m)
	}
	r.F = func(req *http.Request) bool {
		r.mu.RLock()
		defer r.mu.RUnlock()
		host, _, _ := strings.Cut(req.RemoteAddr, ":")
		return !slices.ContainsFunc(r.dropMatcher, func(m Matcher) bool {
			return m.MatchStr(host)
		})
	}
}

func (r *Rule) buildDropFilter() {
	r.acceptMatcher = make([]Matcher, 0, len(r.AcceptHosts))
	for _, h := range r.AcceptHosts {
		m, err := GenMatcher(h)
		if err != nil {
			slog.Error("failed to generate mathcer from host, skip", "raw", h, "err", err)
		}
		r.acceptMatcher = append(r.acceptMatcher, m)
	}
	r.F = func(req *http.Request) bool {
		host, _, _ := strings.Cut(req.RemoteAddr, ":")
		return slices.ContainsFunc(r.acceptMatcher, func(m Matcher) bool {
			return m.MatchStr(host)
		})
	}
}

func (r *Rule) AddAccept(s string) error {
	m, err := GenMatcher(s)
	if err != nil {
		return fmt.Errorf("error while generating new matcher: %s, raw_ip: %s", err, s)
	}

	r.mu.Lock()
	r.acceptMatcher = append(r.acceptMatcher, m)
	r.AcceptHosts = append(r.AcceptHosts, s)
	r.mu.Unlock()

	slog.Info("add new accept matcher", "ip", s)

	return nil
}

func (r *Rule) AddDrop(s string) error {
	m, err := GenMatcher(s)
	if err != nil {
		return fmt.Errorf("error while generating new matcher: %s, raw_ip: %s", err, s)
	}

	r.mu.Lock()
	r.dropMatcher = append(r.dropMatcher, m)
	r.DropHosts = append(r.DropHosts, s)
	r.mu.Unlock()

	slog.Info("add new drop matcher", "ip", s)

	return nil
}

// SafeFilter support mutex for hot reload
// you should always use this instead of raw [Rule.F]
func (r *Rule) SafeFilter(req *http.Request) bool {
	r.mu.RLock()

	defer r.mu.RUnlock()

	return r.F(req)
}

func NewRule(conf io.Reader) (*Rule, error) {
	byteData, err := io.ReadAll(conf)
	if err != nil {
		return nil, fmt.Errorf("while reading configuration data: %s", err)
	}

	var r Rule

	err = json.Unmarshal(byteData, &r)
	if err != nil {
		log.Fatal("error while marshal config, fix code", err)
	}

	r.mu = sync.RWMutex{}

	r.buildFilter()

	return &r, nil
}
