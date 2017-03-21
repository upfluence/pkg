package http_thrift

import "net/http"

type THTTPRequest struct {
	w    *http.ResponseWriter
	r    *http.Request
	lock chan bool
}

func (d *THTTPRequest) Open() error {
	return nil
}

func (d *THTTPRequest) IsOpen() bool {
	return true
}

func (d *THTTPRequest) Close() error {
	return nil
}

func (d *THTTPRequest) Read(buf []byte) (int, error) {
	return d.r.Body.Read(buf)
}

func (d *THTTPRequest) Write(buf []byte) (int, error) {
	return (*d.w).Write(buf)
}

func (d *THTTPRequest) Flush() error {
	d.lock <- true

	return nil
}
