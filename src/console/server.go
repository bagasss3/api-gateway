package console

import (
	"api-gateway/src/config"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run server",
	Long:  "Start running the server",
	Run:   server,
}

func init() {
	RootCmd.AddCommand(serverCmd)
}

func server(cmd *cobra.Command, args []string) {
	addr := config.Port()
	log.Info("Starting API Gateway...")

	http.HandleFunc("/", handleRequest)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var backendServices = map[string]string{
		"cakes":    config.CakeService(),
		"products": config.Servis2(),
	}

	serviceName := parseServiceName(r.URL.Path)
	// urlEndpoint := parseURLEndpoint(r.URL.Path)
	backendURL, ok := backendServices[serviceName]
	if !ok {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	parsedBackendURL, err := parseBackendURL(backendURL)
	if err != nil {
		http.Error(w, "Error parsing backend URL", http.StatusInternalServerError)
		return
	}

	log.Info(fmt.Sprintf("Incoming request to %s%s with method %s", parsedBackendURL, r.URL, r.Method))
	proxy := httputil.NewSingleHostReverseProxy(parsedBackendURL) // req to service
	proxy.ServeHTTP(w, r)
}

func parseServiceName(path string) string {
	//localhost:8080/api/v1/users/login
	// ekstrak users
	parts := strings.Split(path, "/")
	// for _, part := range parts {
	// 	if part != "" && backendServices[part] != "" {
	// 		return part
	// 	}
	// }
	if len(parts) < 2 {
		return "default"
	}
	return parts[2]
}

func parseBackendURL(backendURL string) (*url.URL, error) {
	baseURL, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}
	return baseURL, nil
}
