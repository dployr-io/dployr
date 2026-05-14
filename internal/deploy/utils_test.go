// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"strings"
	"testing"
)

func TestBuildAuthUrl_AlreadyHasCredentials(t *testing.T) {
	url := "https://user:pass@github.com/org/repo.git"
	got, err := buildAuthUrl(url, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != url {
		t.Errorf("buildAuthUrl(already credentialed) = %q, want unchanged", got)
	}
}

func TestBuildAuthUrl_EmptyToken(t *testing.T) {
	url := "https://github.com/org/repo.git"
	got, err := buildAuthUrl(url, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != url {
		t.Errorf("buildAuthUrl(empty token) = %q, want unchanged", got)
	}
}

func TestBuildAuthUrl_HttpNormalisedToHttps(t *testing.T) {
	got, err := buildAuthUrl("http://github.com/org/repo.git", "ghp_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, "https://") {
		t.Errorf("buildAuthUrl(http://) = %q, expected https:// prefix", got)
	}
	if !strings.Contains(got, "ghp_abc") {
		t.Errorf("buildAuthUrl(http://) = %q, expected token injected", got)
	}
}

func TestBuildAuthUrl_GitHub(t *testing.T) {
	got, err := buildAuthUrl("https://github.com/org/repo.git", "ghp_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://x-access-token:ghp_token@github.com/org/repo.git"
	if got != want {
		t.Errorf("buildAuthUrl(github) = %q, want %q", got, want)
	}
}

func TestBuildAuthUrl_GitLab(t *testing.T) {
	got, err := buildAuthUrl("https://gitlab.com/org/repo.git", "glpat_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "oauth2:glpat_token@") {
		t.Errorf("buildAuthUrl(gitlab) = %q, expected oauth2 username", got)
	}
}

func TestBuildAuthUrl_Bitbucket(t *testing.T) {
	got, err := buildAuthUrl("https://bitbucket.org/org/repo.git", "bb_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "x-token-auth:bb_token@") {
		t.Errorf("buildAuthUrl(bitbucket) = %q, expected x-token-auth username", got)
	}
}

func TestBuildAuthUrl_UnknownHttpsHost(t *testing.T) {
	got, err := buildAuthUrl("https://git.internal.company/repo.git", "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "oauth2:tok@") {
		t.Errorf("buildAuthUrl(unknown host) = %q, expected oauth2 fallback", got)
	}
}

func TestBuildAuthUrl_SSHUrl(t *testing.T) {
	url := "git@github.com:org/repo.git"
	got, err := buildAuthUrl(url, "ghp_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != url {
		t.Errorf("buildAuthUrl(SSH) = %q, want unchanged %q", got, url)
	}
}
