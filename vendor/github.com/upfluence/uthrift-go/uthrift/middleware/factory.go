package middleware

import "github.com/upfluence/thrift/lib/go/thrift"

type Factory interface {
	GetMiddleware(string, string) thrift.TMiddleware
}
