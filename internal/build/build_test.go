package build

import "testing"

func Test_updateVersion(t *testing.T) {
	lines := []string{
		"",
		"foo.com/bar/baz v1.0.0",
		"github.com/zendesk/go-generics/bar v99.99.11111",
		"github.com/zendesk/go-generics/baz v99.99.11111",
		"asdf",
		"",
	}

	result := updateVersion(lines, "v2.0.0")

	expected := []string{
		"",
		"foo.com/bar/baz v1.0.0",
		"github.com/zendesk/go-generics/bar v2.0.0",
		"github.com/zendesk/go-generics/baz v2.0.0",
		"asdf",
		"",
	}

	for i, line := range result {
		if line != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], line)
		}
	}
}
