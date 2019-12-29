package docker

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		body          string
		expectedError bool
	}{
		{"docker", false},
	}

	for _, test := range tests {
		c := caddy.NewTestController("dns", test.body)
		c.ServerBlockKeys = []string{"domain.com.:8053", "dynamic.domain.com.:8053"}
		if err := setup(c); (err == nil) == test.expectedError {
			t.Errorf("Unexpected errors: %v", err)
		}
	}
}
