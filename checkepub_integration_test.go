//go:build integration

package checkepub_test

import (
	"testing"

	"github.com/bitfield/checkepub"

	"github.com/google/go-cmp/cmp"
)

func TestAPIReturnsValidForValidFile(t *testing.T) {
	t.Parallel()
	want := ValidResult
	got, err := checkepub.Check("testdata/valid.epub")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAPIReturnsInvalidForInvalidFile(t *testing.T) {
	t.Parallel()
	want := checkepub.Result{
		Status: checkepub.StatusInvalid,
		Errors: []checkepub.ValidationError{
			"RSC-005, ERROR, [Error while parsing file 'element \"metadata\" incomplete; missing required element \"dc:title\"'.], EPUB/content.opf (8-14)",
		},
	}
	got, err := checkepub.Check("testdata/invalid.epub")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
