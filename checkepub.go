package checkepub

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

type Checker struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewChecker() *Checker {
	return &Checker{
		BaseURL: "http://lint.hametuha.pub/validator",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

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
	return ParseResponse(resp.Body)
}

type Result struct {
	Status ValidationStatus
	Errors []ValidationError
}

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

type ValidationStatus string

const (
	StatusValid   ValidationStatus = "Valid"
	StatusInvalid ValidationStatus = "Invalid"
)

type ValidationError string

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

func ParseResponse(r io.Reader) (Result, error) {
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

func Check(path string) (Result, error) {
	return NewChecker().Check(path)
}
