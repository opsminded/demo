package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/opsminded/api"
	"github.com/opsminded/graphlib"
	"github.com/opsminded/service"

	middleware "github.com/oapi-codegen/nethttp-middleware"
)

func main() {
	ex := service.TestableExtractor{
		FrequencyDuration: 20 * time.Second,
		BaseEdges:         []graphlib.Edge{},
		BaseVertices:      []graphlib.Vertex{},
	}

	buildTest2(&ex)

	// TODO: remove this
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := service.New([]service.Extractor{&ex})
	time.Sleep(5 * time.Second)
	s.SetVertexHealth("AndroidApp_CreditSimulators_AppModule", false)
	s.SetVertexHealth("LoadBalancer_01", false)

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

var app_modules = []string{
	"Cards",
	"Payments",
	"Transfers",
	"Investments",
	"Loans",
	"Financing",
	"Insurance",
	"Profile",
	"Notifications",
	"Chat",
	"Rewards",
	"VirtualCard",
	"Portability",
	"CreditSimulators",
}

func Flip(chance int) bool {
	switch {
	case chance <= 0:
		return false // 0 %: nunca
	case chance >= 100:
		return true // 100 %: sempre
	default:
		return rand.Intn(100) < chance
	}
}

func buildOfic(ex *service.TestableExtractor) {
	for h := 0; h < 5; h++ {

		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Key:       fmt.Sprintf("VMDatabaseServer_%d", h),
			Label:     fmt.Sprintf("VMDatabaseServer %d", h),
			Healthy:   true,
			LastCheck: time.Now().UnixNano(),
		})

		for x := 0; x < 4; x++ {
			ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
				Key:       fmt.Sprintf("Database_%d_%d", h, x),
				Label:     fmt.Sprintf("Database %d%d", h, x),
				Healthy:   true,
				LastCheck: time.Now().UnixNano(),
			})

			ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
				Source: fmt.Sprintf("VMDatabaseServer_%d", h),
				Target: fmt.Sprintf("Database_%d_%d", h, x),
			})
		}
	}

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "Internet",
		Label:     "Internet",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "Android",
		Label:     "Android",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})
	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "Iphone",
		Label:     "Iphone",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})

	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "Android",
		Target: "Internet",
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "Iphone",
		Target: "Internet",
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "APIGateway",
		Label:     "APIGateway",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "Internet",
		Target: "APIGateway",
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "LoadBalancer_01",
		Label:     "LB 01",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "LoadBalancer_02",
		Label:     "LB 02",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "APIGateway",
		Target: "LoadBalancer_01",
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "APIGateway",
		Target: "LoadBalancer_02",
	})

	for x := 0; x < 8; x++ {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Key:       fmt.Sprintf("Cluster_%d", x),
			Label:     fmt.Sprintf("Cluster %d", x),
			Healthy:   true,
			LastCheck: time.Now().UnixNano(),
		})
		ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
			Source: "LoadBalancer_01",
			Target: fmt.Sprintf("Cluster_%d", x),
		})
		ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
			Source: "LoadBalancer_02",
			Target: fmt.Sprintf("Cluster_%d", x),
		})
	}

	for x := 0; x < 20; x++ {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Key:       fmt.Sprintf("Service_%d", x),
			Label:     fmt.Sprintf("svc %d", x),
			Healthy:   true,
			LastCheck: time.Now().UnixNano(),
		})
		for y := 0; y < 8; y++ {
			if Flip(60) {
				ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
					Source: fmt.Sprintf("Cluster_%d", y),
					Target: fmt.Sprintf("Service_%d", x),
				})
			}
		}

		for _, v := range ex.BaseVertices {
			if strings.Contains(v.Key, "VMDatabaseServer_") {
				if Flip(30) {
					ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
						Source: fmt.Sprintf("Service_%d", x),
						Target: v.Key,
					})
				}
			}
		}
	}

	for _, label := range app_modules {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Key:       "AndroidApp_" + label,
			Label:     label,
			Healthy:   true,
			LastCheck: time.Now().UnixNano(),
		})
		ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
			Source: "AndroidApp_" + label,
			Target: "Android",
		})
	}

	for _, label := range app_modules {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Key:       "IphoneApp_" + label,
			Label:     label,
			Healthy:   true,
			LastCheck: time.Now().UnixNano(),
		})
		ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
			Source: "IphoneApp_" + label,
			Target: "Iphone",
		})
	}
}

func buildTest(ex *service.TestableExtractor) {
	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "A",
		Label:     "A",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})
	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "B",
		Label:     "B",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})
	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "C",
		Label:     "C",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})
	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "D",
		Label:     "D",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})
	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Key:       "E",
		Label:     "E",
		Healthy:   true,
		LastCheck: time.Now().UnixNano(),
	})

	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "A",
		Target: "B",
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "A",
		Target: "C",
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "C",
		Target: "D",
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Source: "D",
		Target: "E",
	})
}

func buildTest2(ex *service.TestableExtractor) {
	for a := range 100000 {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Key:       fmt.Sprintf("A_%d", a),
			Label:     fmt.Sprintf("A_%d", a),
			Healthy:   true,
			LastCheck: time.Now().UnixNano(),
		})

		for b := range 5 {
			ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
				Key:       fmt.Sprintf("B_%d_%d", a, b),
				Label:     fmt.Sprintf("B_%d_%d", a, b),
				Healthy:   true,
				LastCheck: time.Now().UnixNano(),
			})

			ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
				Source: fmt.Sprintf("A_%d", a),
				Target: fmt.Sprintf("B_%d_%d", a, b),
			})
		}
	}
}
