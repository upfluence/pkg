# cfg

[![CircleCI](https://circleci.com/gh/upfluence/cfg.svg?style=svg)](https://circleci.com/gh/upfluence/cfg)

Convinient configuration builder. Inject data straight to your
configuration struct from various input source (env, flags, files, you
name it!)

## Provider

Implemented providers:

### Environment

TODO

### Flags

TODO

### JSON input

TODO

### Static

TODO

### Create your own

TODO


## Example

The simpliest way to use upfluence/cfg is:

```go
package main

import (
  "context"
  "fmt"
  "os"

  "github.com/upfluence/cfg"
)

type config struct {
  Arg1 string `env:"ARG_1" flag:"arg1"`
  Arg2 string `env:"ARG_2" flag:"arg2"`
}

func main() {
  var (
    c config

    ctx = context.Background()
  )

  if err := cfg.NewDefaultConfigurator().Populate(ctx, &c); err != nil {
    fmt.Fprintf(os.Stderr, "cannot populate config: %s\n", err.Error())
    os.Exit(1)
  }

  fmt.Printf("Arg1: %s, Arg2: %s\n", c.Arg1, c.Arg2)
}
```

By default only the `env` and `flags` provider are used.

You can now provide configuration to your application through:

* `env`:

```
$ ARG_1=foo ARG_2 ./example
Arg1: foo, Arg2: bar
```

* `flags`:

```
$ ./example --arg1 foo --arg2 bar
Arg1: foo, Arg2: bar
```

* `flags:booleans`:

```
$ ./example --arg1 --no-arg2
Arg1: TRUE, Arg2: FALSE
```

## Roadmap

So far the implementation is pretty minimal. There is two main
kinds of improvment.

### Field parsing

The current implementation parses:
* `string`
* `int`, `int64`
* `bool`
* `time.Duration` using `time.ParseDuration`
* `time.Time` based on the format `2006-01-02T15:04:05`
* sub structs
* slices
* maps

We will implement more types in the future (including floats, other int types etc...)

### Other provider

A few more provider are on the roadmap.

* KV Store: like etcd, zk, consul.
* More file types. We could include YAML and TOML file parsing provider
