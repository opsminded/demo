package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/opsminded/api"
	"github.com/opsminded/service"

	middleware "github.com/oapi-codegen/nethttp-middleware"
)

func main() {
	ex := service.TestableExtractor{
		FrequencyDuration: time.Second,
		Edges: []service.Edge{
			{
				Label:       "AB",
				Class:       "DEFAULT",
				Source:      "A",
				Destination: "B",
			},
			{
				Label:       "BC",
				Class:       "DEFAULT",
				Source:      "B",
				Destination: "C",
			},
		},
		Vertices: []service.Vertex{
			{
				Label: "A",
				Class: "DEFAULT",
			},
			{
				Label: "B",
				Class: "DEFAULT",
			},
			{
				Label: "C",
				Class: "DEFAULT",
			},
		},
	}

	s := service.New([]service.Extractor{&ex})
	s.Extract()

	demo := api.New(s)

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	handler := api.NewStrictHandler(demo, []api.StrictMiddlewareFunc{})

	r := http.NewServeMux()

	api.HandlerFromMux(handler, r)

	h := cors(r)
	h = middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: NewAuthenticator(),
		},
	})(h)

	server := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(server.ListenAndServe())
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func NewAuthenticator() openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		return nil
	}
}
