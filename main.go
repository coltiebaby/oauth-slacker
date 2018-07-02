package main

import (
	"github.com/coltiebaby/oauth-slacker/oauth"
	"net/http"
)

func main() {
	s := oauth.NewSlack()
	http.HandleFunc("/api/request-token", s.RequestHandler)
	http.HandleFunc("/api/get-token", s.ResponseHandler)

	http.ListenAndServe(":8080", nil)
}
