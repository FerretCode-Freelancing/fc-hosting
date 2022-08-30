package main

import (
	//"bytes"
	//"encoding/json"
	"errors"
	"fmt"
	//"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)

		w.Write([]byte("hi"))
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		service, err := getService(r.URL.String())

		if err != nil {
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
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
	})

	http.ListenAndServe(":3000", r)
}

type Redirect struct {
	Url string `json:"url"`
}

type RedirectResponse struct {
	Location string
}

func ReverseProxy(address *url.URL) *httputil.ReverseProxy {
	p := httputil.NewSingleHostReverseProxy(address)

	p.Director = func(request *http.Request) {
		request.Host = address.Host
		request.URL.Scheme = address.Scheme
		request.URL.Host = address.Host
		request.URL.Path = address.Path
	}

	/*p.ModifyResponse = func(response *http.Response) error {
		body, _ := io.ReadAll(response.Body)
		request := Redirect{}
		json.Unmarshal(body, &request)

		if request.Url != "" {
			response.StatusCode = http.StatusFound
			response.Header.Set("Content-Type", "application/json")

			redirect := RedirectResponse{request.Url}
			encoded, err := json.Marshal(redirect)

			if err != nil {
				response.Body = io.NopCloser(
					bytes.NewReader([]byte("Couldn't encode redirect")),
				)
				return errors.New("Couldn't encode redirect")
			}

			response.Body = io.NopCloser(
				bytes.NewBuffer(encoded),
			)

			response.Body.Close()
			fmt.Println(response.Body)
		}

		return nil
	}*/

	return p
}

func getService(path string) (string, error) {
	service := strings.Split(path[1:], "/")[0]
	dns := fmt.Sprintf("fc-%s.default.svc.cluster.local", service)

	ips, err := net.LookupIP(dns)

	if err != nil {
		return "", errors.New("Address not found")
	}

	ip := ips[0].String()

	address := fmt.Sprintf("http://%s:3000%s", ip, path)

	return address, nil
}
