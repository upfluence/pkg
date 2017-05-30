package http_thrift

import "net/http"

type HTTPHandler struct {
	server *THTTPServer
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &THTTPRequest{&w, r, make(chan bool)}
	h.server.deliveries <- req

	<-req.lock
}
