package rule

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type IP struct {
	raw  string
	octs []uint8
}

func (ip *IP) Str() string {
	return ip.raw
}

func (ip *IP) Clone() IP {
	return IP{
		raw:  ip.raw,
		octs: slices.Clone(ip.octs),
	}
}

// IPLesser return p1 < p2
func IPLesser(p1, p2 *IP) bool {
	for i := range 4 {
		if p1.octs[i] < p2.octs[i] {
			return true
		} else if p1.octs[i] > p2.octs[i] {
			return false
		}
	}
	return false
}

// IPGreater return p1 > p2
func IPGreater(p1, p2 *IP) bool {
	for i := range 4 {
		if p1.octs[i] > p2.octs[i] {
			return true
		} else if p1.octs[i] < p2.octs[i] {
			return false
		}
	}

	return false
}

// IPEqual return p1 = p2
func IPEqual(p1, p2 *IP) bool {
	for i := range len(p1.octs) {
		if p1.octs[i] != p2.octs[i] {
			return false
		}
	}
	return true
}

func (ip *IP) Match(other *IP) bool {
	return IPEqual(ip, other)
}

func (ip *IP) MatchStr(other string) bool {
	otherIP, err := NewIP(other)
	if err != nil {
		return false
	}
	return IPEqual(ip, otherIP)
}

// func (ip *IP) equal(other *IP) bool {
// 	for i := range len(ip.octs) {
// 		if ip.octs[i] != other.octs[i] {
// 			return false
// 		}
// 	}
//
// 	return true
// }

// func (ip *IP) greater(other *IP) bool {
// 	for i := range len(ip.octs) {
// 		if ip.octs[i] > other.octs[i] {
// 			return true
// 		}
// 	}
//
// 	return false
// }
//
// func (ip *IP) lesser(other *IP) bool {
// 	for i := range len(ip.octs) {
// 		if ip.octs[i] < other.octs[i] {
// 			return true
// 		}
// 	}
//
// 	return false
// }

func NewIP(host string) (*IP, error) {
	host = strings.TrimSpace(host)
	parts := strings.Split(host, ".")

	octs := make([]uint8, 0, 4)
	for _, s := range parts {
		i, _ := strconv.Atoi(s)
		octs = append(octs, uint8(i))
	}

	if len(octs) != 4 {
		return nil, fmt.Errorf("bad ip addr: %s", host)
	}
	return &IP{raw: host, octs: octs}, nil
}

type Matcher interface {
	// interface for support ip and net range
	// return true if ip match
	Match(ip *IP) bool

	// MatchStr receive the dotted decimal format ip as input
	MatchStr(ip string) bool
}

type NetRange struct {
	start *IP
	end   *IP
}

// Match judge if this ip addr in this range
// it can equal to start or end
func (nr *NetRange) Match(ip *IP) bool {
	return !IPGreater(nr.start, ip) && !IPLesser(nr.end, ip)
}

func (nr *NetRange) MatchStr(ip string) bool {
	other, err := NewIP(ip)
	if err != nil {
		return false
	}
	return nr.Match(other)
}

func NewNetRange(start, end *IP) *NetRange {
	return &NetRange{start: start, end: end}
}

// ParseNetRange receive a string with format dotted decimal
// i.e. 192.168.1.1-192.168.1.5   the - must contains
func ParseNetRange(s string) (start *IP, end *IP, err error) {
	before, after, found := strings.Cut(s, "-")
	if !found {
		return nil, nil, fmt.Errorf("sep - not found on input, input: %s", s)
	}

	fstart, err := NewIP(before)
	if err != nil {
		return nil, nil, err
	}

	fend, err := NewIP(after)
	if err != nil {
		return nil, nil, err
	}

	if IPGreater(fstart, fend) {
		return fend, fstart, err
	}

	start = fstart
	end = fend

	return
}

func GenMatcher(s string) (Matcher, error) {
	if strings.Contains(s, "-") {
		start, end, err := ParseNetRange(s)
		if err != nil {
			return nil, err
		}
		return &NetRange{start: start, end: end}, nil
	}

	return NewIP(s)
}
