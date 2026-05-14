// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package traffic

import (
	"math"
	"testing"
)

func TestSubnet24_ValidIPv4(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"192.168.1.100", "192.168.1"},
		{"10.0.0.1", "10.0.0"},
		{"8.8.8.8", "8.8.8"},
		{"255.255.255.0", "255.255.255"},
	}
	for _, c := range cases {
		got := subnet24(c.in)
		if got != c.want {
			t.Errorf("subnet24(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSubnet24_Unparseable(t *testing.T) {
	raw := "not-an-ip"
	got := subnet24(raw)
	if got != raw {
		t.Errorf("subnet24(%q) = %q, want input unchanged", raw, got)
	}
}

func TestSubnet24_IPv6NoPanic(t *testing.T) {
	// IPv6 addresses should not panic and should return something non-empty.
	got := subnet24("2001:db8::1")
	if got == "" {
		t.Error("subnet24(IPv6) returned empty string")
	}
}

func TestStripQuery_WithQuery(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"/path?foo=bar", "/path"},
		{"/a/b/c?x=1&y=2", "/a/b/c"},
		{"/?", "/"},
	}
	for _, c := range cases {
		got := stripQuery(c.in)
		if got != c.want {
			t.Errorf("stripQuery(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestStripQuery_WithoutQuery(t *testing.T) {
	cases := []string{"/path", "/", "/a/b/c"}
	for _, uri := range cases {
		got := stripQuery(uri)
		if got != uri {
			t.Errorf("stripQuery(%q) = %q, want unchanged", uri, got)
		}
	}
}

func TestCadenceCV_FewerThanTwoSamples(t *testing.T) {
	if got := cadenceCV(nil); got != 1.0 {
		t.Errorf("cadenceCV(nil) = %v, want 1.0", got)
	}
	if got := cadenceCV([]float64{1.0}); got != 1.0 {
		t.Errorf("cadenceCV(single) = %v, want 1.0", got)
	}
}

func TestCadenceCV_PerfectlyRegular(t *testing.T) {
	// Evenly spaced — stddev = 0, CV = 0.
	ts := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	got := cadenceCV(ts)
	if got != 0.0 {
		t.Errorf("cadenceCV(regular) = %v, want 0.0", got)
	}
}

func TestCadenceCV_Irregular(t *testing.T) {
	// Highly irregular gaps → CV should be well above the 0.2 bot threshold.
	ts := []float64{1.0, 1.1, 5.0, 5.05, 20.0}
	got := cadenceCV(ts)
	if got <= 0.2 {
		t.Errorf("cadenceCV(irregular) = %v, expected > 0.2", got)
	}
}

func TestCadenceCV_MeanZeroGuard(t *testing.T) {
	// All timestamps identical → all gaps are 0 → mean=0; should return 0, not NaN/panic.
	ts := []float64{5.0, 5.0, 5.0}
	got := cadenceCV(ts)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Errorf("cadenceCV(zero gaps) = %v, expected finite value", got)
	}
	if got != 0.0 {
		t.Errorf("cadenceCV(zero gaps) = %v, want 0.0", got)
	}
}
