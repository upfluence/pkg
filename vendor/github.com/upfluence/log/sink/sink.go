package sink

import "github.com/upfluence/log/record"

type Sink interface {
	Log(record.Record) error
}
