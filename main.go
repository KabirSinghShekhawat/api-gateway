package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	GatewayAddress string `yaml:"gateway_address"`
	ServiceAddress string `yaml:"service_address"`
}

func responseReader(r io.Reader) ([]byte, error) {
	data := make([]byte, 100)
	chunkSize := 100
	for {
		buffer := make([]byte, chunkSize)
		n, err := r.Read(buffer)

		if err != nil && err != io.EOF {
			return nil, err
		}

		data = append(data, buffer...)

		if n > 0 {
			fmt.Printf("read %d bytes\n", n)
		}

		if err == io.EOF {
			break
		}
	}
	return data, nil
}

func main() {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading 'config.yaml' file: %v", err)
	}
	var config Config
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Fatalf("Error unmarshalling 'config.yaml' file: %v", err)
	}

	gatewayServerMux := http.NewServeMux()
	serviceServerMux := http.NewServeMux()

	gatewayServerMux.HandleFunc("/gateway/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		resp, err := http.Get("http://" + config.ServiceAddress + "/service")
		if err != nil {
			fmt.Println("Got an error while calling service!")
			w.Write([]byte(err.Error()))
			return
		}
		var builder strings.Builder
		fmt.Printf("URL: %s\n", r.URL.Path)
		data, err := responseReader(resp.Body)
		if err != nil {
			data = append(data, []byte(err.Error())...)
			builder.Write(data)
		} else {
			builder.WriteString("Content-Length: ")
			builder.WriteString(strconv.Itoa(len(data)))
			builder.WriteString("\n")
			builder.WriteString("Content: ")
			builder.Write(data)
			builder.WriteString("\n")
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(builder.String()))
	})

	fmt.Printf("Gateway address: %s\n", config.GatewayAddress)
	fmt.Printf("Service address: %s\n", config.ServiceAddress)
	fmt.Printf("Listening on %s\n", config.GatewayAddress)

	go func() {
		serviceServerMux.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Service Got Request")
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Hello from Service.\n"))
		})
		log.Fatal(http.ListenAndServe(config.ServiceAddress, serviceServerMux))
	}()
	log.Fatal(http.ListenAndServe(config.GatewayAddress, gatewayServerMux))
}
