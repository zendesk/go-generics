package functions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zendesk/go-generics/functions/internal/test"
)

func TestMergeErrors(t *testing.T) {
	err1 := fmt.Errorf("Error1")
	err2 := fmt.Errorf("Error2")
	err3 := fmt.Errorf("Error3")
	err4 := fmt.Errorf("Error4")

	mergedErr := MergeErrors(err1, err2, err3, err4)
	containsAllErrs := strings.Contains(mergedErr.Error(), err1.Error()) &&
		strings.Contains(mergedErr.Error(), err2.Error()) &&
		strings.Contains(mergedErr.Error(), err3.Error()) &&
		strings.Contains(mergedErr.Error(), err4.Error())

	test.CheckOk(containsAllErrs, "Merged error does not contain all sub-errs!", t)

	var noErrs = []error{}
	errs := MergeErrors(noErrs...)

	test.CheckOk(errs == nil, "errs should be nil but was not", t)
}

func TestHasNonNilError(t *testing.T) {
	err1 := fmt.Errorf("err1")
	err2 := fmt.Errorf("err1")
	err3 := fmt.Errorf("err1")
	var err4 error = nil

	foundNil := HasNonNilError([]error{err1, err2, err3, err4})
	noNilFound := HasNonNilError([]error{err1, err2, err3})
	allNil := HasNonNilError([]error{nil, nil})

	test.CheckOk(foundNil, "Nil error was not detected but should have been", t)
	test.CheckOk(noNilFound, "Nil error was detected but should not have been", t)
	test.CheckOk(!allNil, "Non nil error found but shouldn't have been", t)
}
