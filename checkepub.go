package checkepub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Checker represents a HamePub Lint API client.
type Checker struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewChecker returns a *Checker that can be customised. If you don't need to
// customise the base URL or the HTTP client, you don't need to call NewChecker;
// call Check instead.
func NewChecker() *Checker {
	return &Checker{
		BaseURL: "http://lint.hametuha.pub/validator",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// Check takes a string representing the path to an EPUB file on disk, submits
// it to the HamePub Lint API, and returns a Result indicating whether the file
// is valid, and if not, what validation errors were found. It returns error if
// the HTTP request cannot be constructed or made, or if the HTTP response
// status is anything other than 200 OK.∑
func (c *Checker) Check(epubPath string) (Result, error) {
	f, err := os.Open(epubPath)
	if err != nil {
		return Result{}, err
	}
	defer f.Close()
	body := Base64EncodeReader(f)
	req, err := http.NewRequest(http.MethodPost, c.BaseURL, body)
	if err != nil {
		return Result{}, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, ErrUnexpectedHTTPStatus{
			err: fmt.Errorf("unexpected HTTP response status %q", resp.Status),
		}
	}
	return ParseResponseBody(resp.Body)
}

// Result represents the result of validating an EPUB file. ValidationStatus
// indicates whether or not the file is valid. If it's not, then Errors contains
// a list of the validation errors that were found.
type Result struct {
	Status ValidationStatus
	Errors []ValidationError
}

// String returns a user-friendly string representing the information in the
// Result struct.
func (r Result) String() string {
	if r.Status == StatusValid {
		return "OK"
	}
	s := "Invalid:"
	for _, msg := range r.Errors {
		s += "\n" + string(msg)
	}
	return s
}

// ValidationStatus is used in the Result struct to indicate whether or not the
// file is valid.
type ValidationStatus string

const (
	StatusValid   ValidationStatus = "Valid"
	StatusInvalid ValidationStatus = "Invalid"
)

// ValidationError represents a validation error returned by the API. These are
// the raw errors returned by EPUBCheck, for example:
//
// "PKG-008, FATAL, [Unable to read file 'error in opening zip file'.],
// epubJae5gW.epub".
type ValidationError string

// Base64EncodeReader transforms a supplied io.Reader into one that is
// base64-encoded. For example, if the supplied reader is a file, then the
// returned reader will produce the base64-encoded contents of the file. This is
// used to base64-encode the EPUB file into the HTTP request body without
// slurping the whole file into memory at once.
func Base64EncodeReader(source io.Reader) io.Reader {
	pReader, pWriter := io.Pipe()
	encoder := base64.NewEncoder(base64.StdEncoding, pWriter)
	go func() {
		_, err := io.Copy(encoder, source)
		encoder.Close()
		if err != nil {
			pWriter.CloseWithError(err)
		} else {
			pWriter.Close()
		}
	}()
	return pReader
}

// ParseResponseBody takes an io.Reader containing some JSON-encoded response
// data from the API, and attempts to decode it and produce a corresponding
// Result struct, or an error if the data cannot be read or does not fit the
// expected schema.
func ParseResponseBody(r io.Reader) (Result, error) {
	result := Result{
		Status: StatusValid,
		Errors: []ValidationError{},
	}
	var apiResponse struct {
		Success  bool
		Messages []string
	}
	err := json.NewDecoder(r).Decode(&apiResponse)
	if err != nil {
		return Result{}, err
	}
	if !apiResponse.Success {
		result.Status = StatusInvalid
		for _, msg := range apiResponse.Messages {
			result.Errors = append(result.Errors, ValidationError(msg))
		}
	}
	return result, nil
}

// Check takes a string representing the path to an EPUB file and calls the
// HamePub Lint API to validate the file. This is a convenience wrapper around
// NewChecker.Check for when you don't need to customise anything about the API
// client.
func Check(path string) (Result, error) {
	return NewChecker().Check(path)
}

// ErrUnexpectedHTTPStatus is returned by Check when the HamePub Lint API
// responds with anything other than "200 OK" status. The wrapped error contains
// the unexpected response status.
type ErrUnexpectedHTTPStatus struct {
	err error
}

// Error returns the contents of the wrapped error.
func (e ErrUnexpectedHTTPStatus) Error() string {
	return e.err.Error()
}
