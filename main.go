package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func withCORS(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers.Add("Access-Control-Allow-Origin", "*")
	if len(r.Header.Get("Access-Control-Request-Method")) != 0 {
		headers.Add("access-control-allow-methods", r.Header.Get("Access-Control-Request-Method"))
		r.Header.Del("Access-Control-Request-Method")
	}
	if len(r.Header.Get("Access-Control-Request-Method")) != 0 {
		headers.Add("Access-Control-Request-Method", r.Header.Get("Access-Control-Request-Method"))
		r.Header.Del("Access-Control-Request-Method")
	}
	if len(r.Header.Get("access-control-request-headers")) != 0 {
		headers.Add("access-control-allow-headers", r.Header.Get("access-control-request-headers"))
		r.Header.Del("access-control-request-headers")
	}

	aceh := ""
	for name, _ := range headers {
		// Loop over all values for the name.
		aceh += name + ","
	}
	headers.Add("access-control-expose-headers", aceh)
}

func main() {
	r := gin.Default()

	//Create a catchall route
	r.Any("", handleProxy)

	port := ":7001"
	log.Printf("CORS started listening on %s", port)
	r.Run(port)
}

type corsTransport struct {
	referer            string
	origin             string
	credentials        string
	InsecureSkipVerify bool
}

func (t corsTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// Put in the Referer if specified
	if t.referer != "" {
		r.Header.Add("Referer", t.referer)
	}

	// Do the actual request
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	res, err := http.DefaultTransport.RoundTrip(r)

	if err != nil {
		return nil, err
	}

	res.Header.Set("Access-Control-Allow-Origin", t.origin)
	res.Header.Set("Access-Control-Allow-Credentials", t.credentials)

	return res, nil
}

func handleProxy(c *gin.Context) {
	// Check for the User-Agent header

	origin := "*"
	credentials := "true"
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r := c.Request
	w := c.Writer
	if r.Header.Get("User-Agent") == "" {
		http.Error(w, "Missing User-Agent header", http.StatusBadRequest)
		return
	}

	// Get the optional Referer header
	referer := r.URL.Query().Get("referer")
	if referer == "" {
		referer = r.Header.Get("Referer")
	}

	// Get the URL
	urlParam := r.URL.Query().Get("url")
	realUrlParam := ""
	for key, _ := range r.URL.Query() {
		if key != "url" {
			realUrlParam += "&&" + key + "=" + r.URL.Query().Get(key)
		}
	}

	urlParam += realUrlParam
	// Validate the URL
	urlParsed, err := url.Parse(urlParam)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Check if HTTP(S)
	if urlParsed.Scheme != "http" && urlParsed.Scheme != "https" {
		http.Error(w, "The URL scheme is neither HTTP nor HTTPS", http.StatusBadRequest)
		return
	}

	if r.Method == "OPTIONS" {
		withCORS(w, r)
		return
	}

	InsecureSkipVerify := true
	// Setup for the proxy
	proxy := httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL = urlParsed
			r.Host = urlParsed.Host
		},
		Transport:    corsTransport{referer, origin, credentials, InsecureSkipVerify},
		ErrorHandler: proxyErrorHandler,
	}
	// Execute the request
	proxy.ServeHTTP(w, r)
}

func proxyErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, fmt.Sprintf("%v", err), 500)
}
