package httputil

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/upfluence/pkg/bytesutil"
	"github.com/upfluence/pkg/log"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/uthrift-go/uthrift/context"
)

var bufPool = bytesutil.NewBufferPool()

const headerName = "X-Upfluence-Thrift-Context"

func WithContext(req *http.Request, ctx thrift.Context) *http.Request {
	var buf, err = context.Encode(ctx)

	if err != nil {
		log.Errorf("Can't serialize the context: %v", err)
	} else if len(buf) > 0 {
		req.Header.Set(headerName, base64.StdEncoding.EncodeToString(buf))
	}

	return req.WithContext(ctx)
}

func ExtractContext(req *http.Request) (thrift.Context, func()) {
	if v := req.Header.Get(headerName); v != "" {
		buf := bufPool.Get()
		defer func() { bufPool.Put(buf) }()

		_, err := io.Copy(
			buf,
			base64.NewDecoder(base64.StdEncoding, strings.NewReader(v)),
		)

		if err != nil {
			log.Noticef("Can't decode thrift context header: %s: %s", v, err.Error())
		}

		res, cancel, err := context.Decode(buf.Bytes(), req.Context())

		if err != nil {
			log.Noticef("Can't decode thrift context header: %s: %s", v, err.Error())
		}

		return res, cancel
	}

	res, cancel, _ := context.Decode([]byte{}, req.Context())

	return res, cancel
}
