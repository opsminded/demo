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
		FrequencyDuration: time.Second,
		BaseEdges:         []graphlib.Edge{},
		BaseVertices:      []graphlib.Vertex{},
	}

	for h := 0; h < 5; h++ {

		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Label:   fmt.Sprintf("VMDatabaseServer_%d", h),
			Healthy: true,
		})

		for x := 0; x < 4; x++ {
			ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
				Label:   fmt.Sprintf("Database_%d_%d", h, x),
				Healthy: true,
			})

			ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
				Label: fmt.Sprintf("VMDatabaseServer_%d-Database_%d_%d", h, h, x),
				Source: graphlib.Vertex{
					Label:   fmt.Sprintf("VMDatabaseServer_%d", h),
					Healthy: true,
				},
				Destination: graphlib.Vertex{
					Label:   fmt.Sprintf("Database_%d_%d", h, x),
					Healthy: true,
				},
			})
		}
	}

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Label:   "Internet",
		Healthy: true,
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Label:   "AndroidApp",
		Healthy: true,
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Label: "AndroidApp-Internet",
		Source: graphlib.Vertex{
			Label:   "AndroidApp",
			Healthy: true,
		},
		Destination: graphlib.Vertex{
			Label:   "Internet",
			Healthy: true,
		},
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Label:   "IphoneApp",
		Healthy: true,
	})

	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Label: "IphoneApp-Internet",
		Source: graphlib.Vertex{
			Label:   "IphoneApp",
			Healthy: true,
		},
		Destination: graphlib.Vertex{
			Label:   "Internet",
			Healthy: true,
		},
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Label:   "APIGateway",
		Healthy: true,
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Label: "Internet-APIGateway",
		Source: graphlib.Vertex{
			Label:   "Internet",
			Healthy: true,
		},
		Destination: graphlib.Vertex{
			Label:   "APIGateway",
			Healthy: true,
		},
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Label:   "LoadBalancer_01",
		Healthy: true,
	})

	ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
		Label:   "LoadBalancer_02",
		Healthy: true,
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Label: "APIGateway-LoadBalancer_01",
		Source: graphlib.Vertex{
			Label:   "APIGateway",
			Healthy: true,
		},
		Destination: graphlib.Vertex{
			Label:   "LoadBalancer_01",
			Healthy: true,
		},
	})
	ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
		Label: "APIGateway-LoadBalancer_02",
		Source: graphlib.Vertex{
			Label:   "APIGateway",
			Healthy: true,
		},
		Destination: graphlib.Vertex{
			Label:   "LoadBalancer_02",
			Healthy: true,
		},
	})

	for x := 0; x < 8; x++ {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Label:   fmt.Sprintf("Cluster_%d", x),
			Healthy: true,
		})
		ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
			Label: fmt.Sprintf("LoadBalancer_01-Cluster_%d", x),
			Source: graphlib.Vertex{
				Label:   "LoadBalancer_01",
				Healthy: true,
			},
			Destination: graphlib.Vertex{
				Label:   fmt.Sprintf("Cluster_%d", x),
				Healthy: true,
			},
		})
		ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
			Label: fmt.Sprintf("LoadBalancer_02-Cluster_%d", x),
			Source: graphlib.Vertex{
				Label:   "LoadBalancer_02",
				Healthy: true,
			},
			Destination: graphlib.Vertex{
				Label:   fmt.Sprintf("Cluster_%d", x),
				Healthy: true,
			},
		})
	}

	for x := 0; x < 20; x++ {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Label:   fmt.Sprintf("Service_%d", x),
			Healthy: true,
		})
		for y := 0; y < 8; y++ {
			if Flip(60) {
				ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
					Label: fmt.Sprintf("Cluster_%d-Service_%d", y, x),
					Source: graphlib.Vertex{
						Label:   fmt.Sprintf("Cluster_%d", y),
						Healthy: true,
					},
					Destination: graphlib.Vertex{
						Label:   fmt.Sprintf("Service_%d", x),
						Healthy: true,
					},
				})
			}
		}

		for _, v := range ex.BaseVertices {
			if strings.Contains(v.Label, "VMDatabaseServer_") {
				if Flip(30) {
					ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
						Label: fmt.Sprintf("%s_Service_%d", v.Label, x),
						Source: graphlib.Vertex{
							Label:   fmt.Sprintf("Service_%d", x),
							Healthy: true,
						},
						Destination: graphlib.Vertex{
							Label:   v.Label,
							Healthy: true,
						},
					})
				}
			}
		}
	}

	for _, v := range android_app_modules {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Label:   v,
			Healthy: true,
		})
		ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
			Label: "AndroidApp" + "-" + v,
			Source: graphlib.Vertex{
				Label:   v,
				Healthy: true,
			},
			Destination: graphlib.Vertex{
				Label:   "AndroidApp",
				Healthy: true,
			},
		})
	}

	for _, v := range iphone_app_modules {
		ex.BaseVertices = append(ex.BaseVertices, graphlib.Vertex{
			Label:   v,
			Healthy: true,
		})
		ex.BaseEdges = append(ex.BaseEdges, graphlib.Edge{
			Label: "IphoneApp" + "-" + v,
			Source: graphlib.Vertex{
				Label:   v,
				Healthy: true,
			},
			Destination: graphlib.Vertex{
				Label:   "IphoneApp",
				Healthy: true,
			},
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := service.New(ctx, time.Second, []service.Extractor{&ex}, nil)
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

var android_app_modules = []string{
	"AndroidApp_Cards_AppModule",
	"AndroidApp_Payments_AppModule",
	"AndroidApp_Transfers_AppModule",
	"AndroidApp_Investments_AppModule",
	"AndroidApp_Loans_AppModule",
	"AndroidApp_Financing_AppModule",
	"AndroidApp_Insurance_AppModule",
	"AndroidApp_Profile_AppModule",
	"AndroidApp_Notifications_AppModule",
	"AndroidApp_Chat_AppModule",
	"AndroidApp_Rewards_AppModule",
	"AndroidApp_VirtualCard_AppModule",
	"AndroidApp_Portability_AppModule",
	"AndroidApp_CreditSimulators_AppModule",
}

var iphone_app_modules = []string{
	"IphoneApp_Cards_AppModule",
	"IphoneApp_Payments_AppModule",
	"IphoneApp_Transfers_AppModule",
	"IphoneApp_Investments_AppModule",
	"IphoneApp_Loans_AppModule",
	"IphoneApp_Financing_AppModule",
	"IphoneApp_Insurance_AppModule",
	"IphoneApp_Profile_AppModule",
	"IphoneApp_Notifications_AppModule",
	"IphoneApp_Chat_AppModule",
	"IphoneApp_Rewards_AppModule",
	"IphoneApp_VirtualCard_AppModule",
	"IphoneApp_Portability_AppModule",
	"IphoneApp_CreditSimulators_AppModule",
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
