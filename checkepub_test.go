package checkepub_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bitfield/checkepub"

	"github.com/google/go-cmp/cmp"
)

var (
	ValidResult = checkepub.Result{
		Status: checkepub.StatusValid,
		Errors: []checkepub.ValidationError{},
	}

	InvalidResult = checkepub.Result{
		Status: checkepub.StatusInvalid,
		Errors: []checkepub.ValidationError{
			"PKG-008, FATAL, [Unable to read file 'error in opening zip file'.], epubJae5gW.epub",
			"PKG-003, ERROR, [Unable to read EPUB file header.  This is likely a corrupted EPUB file.], epubJae5gW.epub",
		},
	}
)

func TestBase64EncodeReaderCorrectlyEncodesInput(t *testing.T) {
	t.Parallel()
	input := "Hello, world"
	want := []byte("SGVsbG8sIHdvcmxk")
	r := checkepub.Base64EncodeReader(strings.NewReader(input))
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestNewCheckerReturnsCorrectlyConfiguredChecker(t *testing.T) {
	t.Parallel()
	want := &checkepub.Checker{
		BaseURL: "http://lint.hametuha.pub/validator",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
	got := checkepub.NewChecker()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestCheckCallsAPIWithCorrectB64DataInBody(t *testing.T) {
	t.Parallel()
	input := []byte("Hello, world")
	epubPath := t.TempDir() + "/test.epub"
	err := os.WriteFile(epubPath, input, 0600)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		called = true
		want := []byte("SGVsbG8sIHdvcmxk")
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(want, got) {
			t.Error(cmp.Diff(want, got))
		}
	}))
	defer ts.Close()
	checker := checkepub.NewChecker()
	checker.BaseURL = ts.URL
	checker.HTTPClient = ts.Client()
	_, _ = checker.Check(epubPath)
	if !called {
		t.Error("API not called")
	}
}

func TestCheckerGivesCorrectResultForValidEpubResponse(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("testdata/valid.json")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		io.Copy(w, f)
	}))
	defer ts.Close()
	checker := checkepub.NewChecker()
	checker.BaseURL = ts.URL
	checker.HTTPClient = ts.Client()
	got, err := checker.Check("testdata/dummy")
	if err != nil {
		t.Fatal(err)
	}
	want := ValidResult
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestCheckerGivesCorrectResultForInvalidEpubResponse(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("testdata/badzip.json")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		io.Copy(w, f)
	}))
	defer ts.Close()
	checker := checkepub.NewChecker()
	checker.BaseURL = ts.URL
	checker.HTTPClient = ts.Client()
	got, err := checker.Check("testdata/dummy")
	if err != nil {
		t.Fatal(err)
	}
	want := InvalidResult
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseResponseCorrectForValidEpubResponse(t *testing.T) {
	t.Parallel()
	data, err := os.Open("testdata/valid.json")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	want := ValidResult
	got, err := checkepub.ParseResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseResponseCorrectForInvalidEpubResponse(t *testing.T) {
	t.Parallel()
	data, err := os.Open("testdata/badzip.json")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	want := InvalidResult
	got, err := checkepub.ParseResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStringFormatsValidResultCorrectly(t *testing.T) {
	t.Parallel()
	input := ValidResult
	want := "OK"
	got := input.String()
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestStringFormatsInvalidResultCorrectly(t *testing.T) {
	t.Parallel()
	input := InvalidResult
	want := "Invalid:\nPKG-008, FATAL, [Unable to read file 'error in opening zip file'.], epubJae5gW.epub\nPKG-003, ERROR, [Unable to read EPUB file header.  This is likely a corrupted EPUB file.], epubJae5gW.epub"
	got := input.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
