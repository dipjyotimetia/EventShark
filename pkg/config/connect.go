package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/twmb/franz-go/pkg/kgo"
)

func tlsConnect() {
	// Load the CA certificate
	caCert, err := os.ReadFile("path/to/ca.crt")
	if err != nil {
		log.Fatalf("Failed to load CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Load the client certificate and key
	cert, err := tls.LoadX509KeyPair("path/to/client.crt", "path/to/client.key")
	if err != nil {
		log.Fatalf("Failed to load client certificate and key: %v", err)
	}

	// Create a TLS configuration
	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	// Create a Kafka client with the TLS configuration
	client, err := kgo.NewClient(
		kgo.SeedBrokers("broker1:9092", "broker2:9092"),
		kgo.DialTLSConfig(tlsConfig),
	)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer client.Close()

	// Use the client...
	fmt.Println("Kafka client created successfully")
}
