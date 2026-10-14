# Migrating from `pkg` v1 to `pkg/v2`

`github.com/upfluence/pkg` (v1, last release `v1.14.1`) is deprecated and
unmaintained. All fixes land in `github.com/upfluence/pkg/v2` only.

## 1. Rewrite the import path

```sh
grep -rl 'github.com/upfluence/pkg' --include='*.go' . \
  | xargs sed -i '' -E 's#"github.com/upfluence/pkg(/|")#"github.com/upfluence/pkg/v2\1#'
```

Then update `go.mod`:

```sh
go get github.com/upfluence/pkg/v2@latest
go mod tidy
```

On Linux drop the empty string after `-i`. Make sure no import already ending
in `/v2` gets rewritten twice.

## 2. Removed packages

These v1 packages do not exist in v2.

| v1 package | Replacement |
|---|---|
| `cfg` | renamed to `envutil` |
| `generics` | use `slices`, `maps` and `pointers` |
| `stringutil` batch and pointer helpers | use `slices` and `pointers` |
| `amqp/...` | moved to `github.com/upfluence/amqp` (improved version) |
| `prometheus/...` | dropped |
| `tracing/...` | dropped |
| `httputil` | dropped |
| `stringcache/...` | dropped |
| `cassandrautil`, `testing/cassandra`, `testing/postgres`, `testing/sqlite3` | dropped |
| `testutil` | dropped |
| `multierror` | dropped |
| `container/lru` | see `cache` and `cache/policy` |
| `pool/bounded`, `pool/middleware/logger` | dropped |

## 3. Behavior changes

- `thrift/hasher`: the hasher was rewritten to avoid collisions, so hashes
  differ from v1.
- v2 requires Go 1.24 or newer.

## 4. New in v2

`envutil`, `discovery/balancer/simple`, `discovery/resolver/filter`,
`discovery/resolver/transform`, `thrift/thrifttest`, `syncutil` additions
(`RetryableOnce`, `LazyOnce`), and the `cache` Fetcher/MultiFetcher and
policies.
