package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"

	sum "github.com/vikyd/go-checksum/checksum"
)

// Modules must be in appropriate order based on inter-module dependencies
var modules = []string{
	"test",
	"serialize",
	"datastructures",
	"ratelimit",
	"encryption",
	"functions",
	"cache",
}

const (
	ModulePrefix = "github.com/zendesk/go-generics"
)

var goModChecksums = map[string]string{}
var goDirChecksums = map[string]string{}

var processedModules = []string{}

func main() {
	if len(os.Args) < 2 {
		panic("No version provided. Must be in the format vX.X.X")
	}

	version := os.Args[1]

	// iterate over modules, update go.mod, build, publish
	// replace anything prefixed with go-generics with the next version

	for _, module := range modules {
		// Update go mod to specify new version
		err := updateGoMod(module, version)
		panicOnErr(err)

		// checksum new mod file, add to map
		err = checkSumGoMod(module)
		panicOnErr(err)

		// For each processed module, update checksums in go.sum for this module.
		for _, processedModule := range processedModules {
			fmt.Printf("Module: %s - Updating go.sum go.mod checksum for module %s\n", module, processedModule)
			err = updateGoModCheckSum(module, processedModule, version)
			panicOnErr(err)
			fmt.Printf("Module: %s - Updating go.sum directory checksum for module %s\n", module, processedModule)
			err = updateDirCheckSum(module, processedModule, version)
			panicOnErr(err)
		}

		// checksum new module directory, add to map
		err = checkSumModuleDir(module, fmt.Sprintf("%s/%s@%s", ModulePrefix, module, version))
		panicOnErr(err)

		processedModules = append(processedModules, module)
	}
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
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

// calculate checksum for module post-update and add to map
func checkSumGoMod(module string) error {
	hash, err := sum.HashGoMod(fmt.Sprintf("../%s/go.mod", module))
	if err != nil {
		return err
	}
	goModChecksums[module] = hash.GoCheckSum
	return nil
}

// calculate checksum for module directory post-update and add to map
func checkSumModuleDir(module, version string) error {
	fmt.Printf("Module: %s - Building checksum for version %s\n", module, version)
	hash, err := sum.HashDir(fmt.Sprintf("../%s/", module), version)
	if err != nil {
		return err
	}
	goDirChecksums[module] = hash.GoCheckSum
	return nil
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

func createBackup(fileName string) error {
	backup := fileName + ".bak"
	input, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	err = os.WriteFile(backup, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func updateGoMod(module, version string) error {
	mod := fmt.Sprintf("../%s/go.mod", module)
	modLines, err := readFileLines(mod)
	if err != nil {
		return err
	}

	regex := "(\\s*|require\\s+)github.com/zendesk/go-generics/[a-zA-Z]+\\s+"
	r := regexp.MustCompile(regex)
	updated := []string{}
	for _, line := range modLines {
		// regex match line for "module github.com/zendesk/go-generics"
		if isMatch, _ := regexp.MatchString(regex, line); isMatch {
			linePrefix := r.FindString(line)
			newLine := linePrefix + version
			updated = append(updated, newLine)
		} else {
			updated = append(updated, line)
		}

	}

	return writeLines(updated, mod)
}

func updateGoModCheckSum(module, moduleToSum, version string) error {
	sumFile := fmt.Sprintf("../%s/go.sum", module)
	lines, err := readFileLines(sumFile)
	if err != nil {
		return err
	}

	// regex for go.sum line for module and version
	regex := fmt.Sprintf("\\s*github.com/zendesk/go-generics/%s\\s+v[0-9]+\\.[0-9]+\\.[0-9]+\\/go.mod\\s+", moduleToSum)
	updated := []string{}
	shouldAdd := false
	for _, line := range lines {
		// if any version exists in go.sum for this module, remove it.
		if isMatch, _ := regexp.MatchString(regex, line); isMatch {
			shouldAdd = true
			continue
		} else {
			updated = append(updated, line)
		}
	}

	if shouldAdd {
		newLine := ModulePrefix + "/" + moduleToSum + " " + version + "/go.mod" + " " + goModChecksums[moduleToSum]
		updated = append(updated, newLine)
	}

	// sort lines
	sort.Strings(updated)

	return writeLines(updated, sumFile)
}

func updateDirCheckSum(module, moduleToSum, version string) error {
	sumFile := fmt.Sprintf("../%s/go.sum", module)
	lines, err := readFileLines(sumFile)
	if err != nil {
		return err
	}

	// regex for go.sum line for module
	regex := fmt.Sprintf("\\s*github.com/zendesk/go-generics/%s\\s+v[0-9]+\\.[0-9]+\\.[0-9]+\\s+", moduleToSum)
	updated := []string{}
	shouldAdd := false
	for _, line := range lines {
		// if this any version exists in go.sum for this module, remove it.
		if isMatch, _ := regexp.MatchString(regex, line); isMatch {
			shouldAdd = true
			continue
		} else {
			updated = append(updated, line)
		}

	}

	// add sum for latest version if a prior version was removed, resort file
	if shouldAdd {
		newLine := ModulePrefix + "/" + moduleToSum + " " + version + " " + goDirChecksums[moduleToSum]
		updated = append(updated, newLine)
	}

	sort.Strings(updated)

	return writeLines(updated, sumFile)
}
