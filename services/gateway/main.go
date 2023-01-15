package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(httprate.LimitByIP(50, 1*time.Minute))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	r.Post("/*", func(w http.ResponseWriter, r *http.Request) {
		Request(w, r)
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		Request(w, r)
	})

	r.Patch("/*", func(w http.ResponseWriter, r *http.Request) {
		Request(w, r)
	})

	r.Delete("/*", func(w http.ResponseWriter, r *http.Request) {
		Request(w, r)
	})

	http.ListenAndServe(":3000", r)
}

type Redirect struct {
	Url string `json:"url"`
}

type RedirectResponse struct {
	Location string
}

func Request(w http.ResponseWriter, r *http.Request) {
	service, err := getService(r.URL.String())

	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	}

	if proxyUrl, err := url.Parse(service); err != nil {
		fmt.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	} else {
		ReverseProxy(proxyUrl).ServeHTTP(w, r)
	}
}

func ReverseProxy(address *url.URL) *httputil.ReverseProxy {
	p := httputil.NewSingleHostReverseProxy(address)

	p.Director = func(request *http.Request) {
		request.Host = address.Host
		request.URL.Scheme = address.Scheme
		request.URL.Host = address.Host
		request.URL.Path = address.Path
	}

	return p
}

func getService(path string) (string, error) {
	services := strings.Split(path[1:], "/")
	service := services[0]

	if service == "api" {
		service = services[1]
	}

	dns := fmt.Sprintf("fc-%s.default.svc.cluster.local", service)

	ips, err := net.LookupIP(dns)

	if err != nil {
		return "", errors.New("Address not found")
	}

	ip := ips[0].String()

	address := fmt.Sprintf("http://%s:3000%s", ip, path)

	return address, nil
}
