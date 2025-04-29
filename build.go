package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

// Modules must be in appropriate order based on inter-module dependencies
var modules = []string{
	"serialize",
	"datastructures",
	"ratelimit",
	"encryption",
	"functions",
	"cache",
}

func main() {

	if len(os.Args) < 2 {
		panic("No version provided. Must be in the format vX.X.X")
	}

	version := os.Args[1]

	// iterate over modules, update go.mod, build, publish
	// replace anything prefixed with go-generics with the next version

	for _, module := range modules {
		fileName := fmt.Sprintf("%s/go.mod", module)
		lines, err := readFileLines(fileName)
		if err != nil {
			panic(err)
		}

		updated := updateVersion(lines, version)

		backup := fmt.Sprintf("%s/go.mod.bak", module)
		err = writeLines(lines, backup)
		if err != nil {
			panic(err)
		}

		// re-write file
		err = writeLines(updated, fileName)
		if err != nil {
			panic(err)
		}
	}

}

func readFileLines(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// writes the provided lines to the file by overwriting the file
func writeLines(lines []string, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

func updateVersion(lines []string, version string) []string {
	regex := "(\\s*|require\\s+)github.com/zendesk/go-generics/[a-zA-Z]+\\s+"
	r := regexp.MustCompile(regex)
	updated := []string{}
	for _, line := range lines {
		// regex match line for "module github.com/zendesk/go-generics"
		if isMatch, _ := regexp.MatchString(regex, line); isMatch {
			linePrefix := r.FindString(line)
			newLine := linePrefix + version
			updated = append(updated, newLine)
		} else {
			updated = append(updated, line)
		}

	}

	return updated
}
