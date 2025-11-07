package deploy

import (
	"dployr/pkg/shared"
	"testing"
)

func TestBuildAuthUrl(t *testing.T) {
	config := &shared.Config{
		GitHubToken:    "ghp_test123",
		GitLabToken:    "glpat_test456",
		BitBucketToken: "bb_test789",
	}

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "github with token",
			url:      "https://github.com/user/repo.git",
			expected: "https://git:ghp_test123@github.com/user/repo.git",
		},
		{
			name:     "gitlab with token",
			url:      "https://gitlab.com/user/repo.git",
			expected: "https://git:glpat_test456@gitlab.com/user/repo.git",
		},
		{
			name:     "bitbucket with token",
			url:      "https://bitbucket.org/user/repo.git",
			expected: "https://git:bb_test789@bitbucket.org/user/repo.git",
		},
		{
			name:     "already authenticated",
			url:      "https://user:pass@github.com/user/repo.git",
			expected: "https://user:pass@github.com/user/repo.git",
		},
		{
			name:     "unknown provider",
			url:      "https://example.com/user/repo.git",
			expected: "https://example.com/user/repo.git",
		},
		{
			name:     "http converted to https",
			url:      "http://github.com/user/repo.git",
			expected: "https://git:ghp_test123@github.com/user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildAuthUrl(tt.url, config)
			if err != nil {
				t.Errorf("buildAuthUrl() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("buildAuthUrl() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildAuthUrlNoToken(t *testing.T) {
	config := &shared.Config{}
	url := "https://github.com/user/repo.git"
	
	result, err := buildAuthUrl(url, config)
	if err != nil {
		t.Errorf("buildAuthUrl() error = %v", err)
	}
	if result != url {
		t.Errorf("buildAuthUrl() without token should return original URL, got %q", result)
	}
}
