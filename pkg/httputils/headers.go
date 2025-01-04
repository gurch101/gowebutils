package httputils

import "net/http"

const ContentTypeHeader = "Content-Type"

func SetJSONContentTypeRequestHeader(req *http.Request) {
	req.Header.Set(ContentTypeHeader, "application/json")
}

func SetJSONContentTypeResponseHeader(w http.ResponseWriter) {
	w.Header().Set(ContentTypeHeader, "application/json")
}
