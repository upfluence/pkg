package httputil

import "net/http"

func HealcheckHander(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
