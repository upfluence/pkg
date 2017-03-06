package serializer

import "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/upfluence/thrift/lib/go/thrift"

type TStruct interface {
	Write(thrift.TProtocol) error
	Read(thrift.TProtocol) error
}
