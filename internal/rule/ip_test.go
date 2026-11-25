package rule_test

import (
	"testing"

	"simple-http/filter/internal/rule"
)

func TestIPCompare(t *testing.T) {
	var base *rule.IP
	var err error

	ipS := "192.168.5.2.1"
	base, err = rule.NewIP(ipS)
	if err == nil {
		t.Errorf("fail to report error when encounter invalid ip, ip: %s", ipS)
	}

	ipS = "192.168.1.34"
	base, err = rule.NewIP(ipS)
	if err != nil {
		t.Errorf("error while parse ip: %s", ipS)
	}
	if !base.MatchStr(ipS) {
		t.Errorf("error while compare ip, expect equal: %s", ipS)
	}

	if base.MatchStr("192.168.1.33") {
		t.Errorf("error while mismatch ip, expect false, ip string: %s", ipS)
	}
}

func TestIPRange(t *testing.T) {
	input := "100.1.2.5-180.70.6.1"
	s, e, err := rule.ParseNetRange(input)
	if err != nil {
		t.Errorf("error while parsing input range: %s", input)
	}

	if !s.MatchStr("100.1.2.5") {
		t.Errorf("error while comparing start")
		t.Logf("start parsed as: %s", s.Str())
	}

	if !e.MatchStr("180.70.6.1") {
		t.Errorf("error while comparing end")
		t.Logf("end parsed as: %s, expectd: %s", e.Str(), "180.70.6.1")
	}

	nr := rule.NewNetRange(s, e)
	if !nr.MatchStr("100.5.8.1") {
		t.Errorf("error while match range")
	}

	if !nr.MatchStr("120.81.67.1") {
		t.Errorf("error while match range")
	}

	if nr.MatchStr("100.1.1.3") {
		t.Errorf("error while match range")
	}

	if nr.MatchStr("100.1.2.3") {
		t.Errorf("error while match range")
	}

	input = "180.70.6.1-100.1.2.5"

	s, e, err = rule.ParseNetRange(input)
	if err != nil {
		t.Errorf("error while parsing input range: %s", input)
	}

	if !s.MatchStr("100.1.2.5") {
		t.Errorf("error while comparing start")
		t.Logf("start parsed as: %s", s.Str())
	}

	if !e.MatchStr("180.70.6.1") {
		t.Errorf("error while comparing end")
		t.Logf("end parsed as: %s, expectd: %s", e.Str(), "180.70.6.1")
	}

	nr = rule.NewNetRange(s, e)
	if !nr.MatchStr("100.5.8.1") {
		t.Errorf("error while match range")
	}

	if !nr.MatchStr("120.81.67.1") {
		t.Errorf("error while match range")
	}

	if nr.MatchStr("100.1.1.3") {
		t.Errorf("error while match range")
	}

	if nr.MatchStr("100.1.2.3") {
		t.Errorf("error while match range")
	}
}
