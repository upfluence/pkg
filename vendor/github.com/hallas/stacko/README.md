# stacko [![Build Status](https://travis-ci.org/hallas/stacko.svg?branch=master)](https://travis-ci.org/hallas/stacko) [![GoDoc](https://godoc.org/github.com/hallas/stacko?status.svg)](https://godoc.org/github.com/hallas/stacko) [![Tips](https://img.shields.io/gratipay/hallas.svg)](https://gratipay.com/hallas) [![Join the chat at https://gitter.im/hallas/stacko](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/hallas/stacko?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

The Stacko project generates structured stacktraces for your Go programming
needs.

The general form of a frame is as seen below.

```go
type Frame struct {
  FileName     string
  FunctionName string
  PackageName  string
  Path         string
  LineNumber   int
  InDomain     bool
  PreContext   []string
  PostContext  []string
  Context      string
}
```

Most of the fields are self explanatory. `InDomain` is a boolean that tells you
if the frame is within the same package as the first call. The `Context` fields
are the actual lines of code that precede and procede the context of the frame.
Please note that these context fields are only provided if the source code is
present in the local file system.

## API

### type Stacktrace

```go
type Stacktrace []Frame
```

### func NewStacktrace

```go
func NewStacktrace (skip int) (Stacktrace, error)
```

Returns a new `Stacktrace` skipping the first `skip` frames.

Please refer to the [![GoDoc](https://godoc.org/github.com/hallas/stacko?status.svg)](https://godoc.org/github.com/hallas/stacko)
page for the full API documentation.
