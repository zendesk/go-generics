package build

import (
	"fmt"
	"regexp"
)

func updateVersion(lines []string, version string) []string {
	regex := "\\s*github.com/zendesk/go-generics/[a-zA-Z]+\\s+"
	r := regexp.MustCompile(regex)
	updated := []string{}
	for _, line := range lines {
		fmt.Println(line)
		// regex match line for "module github.com/zendesk/go-generics"
		if isMatch, _ := regexp.MatchString(regex, line); isMatch {
			fmt.Printf("Got match for line: %s\n", line)
			linePrefix := r.FindString(line)
			newLine := linePrefix + version
			updated = append(updated, newLine)
		} else {
			updated = append(updated, line)
		}

	}

	return updated
}
