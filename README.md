[![Go Reference](https://pkg.go.dev/badge/github.com/bitfield/checkepub.svg)](https://pkg.go.dev/github.com/bitfield/checkepub)[![Go Report Card](https://goreportcard.com/badge/github.com/bitfield/checkepub)](https://goreportcard.com/report/github.com/bitfield/checkepub)

```go
import "github.com/bitfield/checkepub"
```

## Introduction

`checkepub` is a library and a CLI tool for validating [EPUB](https://en.wikipedia.org/wiki/EPUB) ebook files against the official [EPUB specification](https://www.iso.org/standard/53255.html).

## Installing the tool

You can install the tool using the `go install` command:

```
go install github.com/bitfield/checkepub/cmd/checkepub@latest
```

## Using the tool

If you want to validate an EPUB file named, for example, `test.epub`, run the following command:

```
checkepub test.epub
```

If the file is valid, you will see the message:

```
OK
```

If there are validation errors, you will see instead, for example:

```
Invalid:
[ list of errors ]
```

and the command will exit with a non-zero exit status. This makes it suitable for running automatically, for example in CI pipelines.

## Using the library

To use `checkepub` in your own Go programs, import the package:

```go
import "github.com/bitfield/checkepub"
```

To check an EPUB file (for example `test.epub`), pass its pathname to the `Check` function:

```go
result, err := checkepub.Check("test.epub")
```

The `result` struct has the following fields:

* `Status`: if the file is valid, this will be `checkepub.StatusValid`, or `checkepub.StatusInvalid` otherwise
* `Errors`: if there are validation errors, which are strings, this slice will contain each of them as a `checkepub.ValidationError`.

## Customisation

To customise either the API URL or the HTTP client used to make the requests, first call `NewChecker`:

```go
checker := checkepub.NewChecker()
```

Then set, for example, the `BaseURL` field:

```go
checker.BaseURL = "https://example.com"
```

or the `HTTPClient` field:

```go
checker.HTTPClient = &http.Client{
    Timeout: 10 * time.Second,
}
```

Then call its `Check` method to use it:

```go
result, err := checker.Check("test.epub")
```

## How it works

`checkepub` is a client for the REST API service provided by [HamePub Lint](https://lint.hametuha.pub/), which ultimately runs the [EPUBCheck tool](https://www.w3.org/publishing/epubcheck/). `checkepub` requires access to the HamePub API server at:

```
http://lint.hametuha.pub/validator
```

Although an HTTPS endpoint is available, its certificate is currently (February 2022) expired, so `checkepub` uses the HTTP endpoint instead. Be aware that this means your EPUB data is sent unencrypted over the wire, so don't use this tool with confidential data.

## Terms of service

The HamePub Lint API's terms of service are:

> This service is provided "AS IS". No warranty, No liability. Everything ows to you.

The `checkepub` tool and library inherits these terms of service.
