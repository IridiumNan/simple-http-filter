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

	"simple-http/filter/internal/filter"
)

type Rule struct {
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
	switch strings.ToLower(r.DefaultPolicy) {
	case "accept":
		r.buildAcceptFilter()
	case "drop":
		r.buildDropFilter()
	default:
		log.Fatalf("invalid policy!!!, policy must be accept or drop, got: %s", r.DefaultPolicy)
	}
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

	r.buildFilter()

	return &r, nil
}
