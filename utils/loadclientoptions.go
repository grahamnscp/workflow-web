package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"strconv"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"

	"github.com/uber-go/tally/v4/prometheus"
	sdktally "go.temporal.io/sdk/contrib/tally"

	dataconverter "webapp/dataconverter"
)

/* LoadClientOptions - Return client options for Temporal Cloud */
func LoadClientOptions(addMetrics bool, metricsPort string) (client.Options, error) {

	// Read env variables
	targetHost := os.Getenv("TEMPORAL_HOST_URL")
	namespace := os.Getenv("TEMPORAL_NAMESPACE")

	// use server mTLS?
	useTLS := false
	useTLS, _ = strconv.ParseBool(os.Getenv("USE_TLS"))
	var cert tls.Certificate
	var clientCert, clientKey, serverName string
	var insecureSkipVerify bool = false
	var serverCAPool *x509.CertPool

	if useTLS {
		clientCert := os.Getenv("TEMPORAL_TLS_CERT")
		clientKey := os.Getenv("TEMPORAL_TLS_KEY")

		// Optional:
		serverRootCACert := os.Getenv("TEMPORAL_SERVER_ROOT_CA_CERT")
		serverName := os.Getenv("TEMPORAL_SERVER_NAME")

		insecureSkipVerify, _ := strconv.ParseBool(os.Getenv("TEMPORAL_INSECURE_SKIP_VERIFY"))

		// Load client cert
		cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return client.Options{}, fmt.Errorf("failed loading client cert and key: %w", err)
		}

		// Load server CA if given
		if serverRootCACert != "" {
			serverCAPool = x509.NewCertPool()
			b, err := os.ReadFile(serverRootCACert)
			if err != nil {
				return client.Options{}, fmt.Errorf("failed reading server CA: %w", err)
			} else if !serverCAPool.AppendCertsFromPEM(b) {
				return client.Options{}, fmt.Errorf("server CA PEM file invalid")
			}
		}
		log.Println("LoadClientOptions: SSL Connection:", clientCert, clientKey, cert, serverRootCACert, serverName, insecureSkipVerify)
	}

	// encrypt payloads?
	encyptPayload, _ := strconv.ParseBool(os.Getenv("ENCRYPT_PAYLOAD"))

	log.Println("LoadClientOptions: Connection", targetHost, namespace, clientCert, clientKey)

	// Return client options
	if encyptPayload {

		if useTLS {
			// Temporal Cloud

			if addMetrics {
				return client.Options{
					HostPort:  targetHost,
					Namespace: namespace,
					ConnectionOptions: client.ConnectionOptions{
						TLS: &tls.Config{
							Certificates:       []tls.Certificate{cert},
							RootCAs:            serverCAPool,
							ServerName:         serverName,
							InsecureSkipVerify: insecureSkipVerify,
						},
					},
					Logger: NewTClientLogger(),

					// Set DataConverter to ensure that workflow inputs and results are
					// encrypted/decrypted as required.
					DataConverter: dataconverter.NewEncryptionDataConverter(
						converter.GetDefaultDataConverter(),
						dataconverter.DataConverterOptions{KeyID: os.Getenv("DATACONVERTER_ENCRYPTION_KEY_ID")},
					),

					// Add SDK Metrics endpoint (for default Go SDK metrics)
					MetricsHandler: sdktally.NewMetricsHandler(newPrometheusScope(
						prometheus.Configuration{
							ListenAddress: "0.0.0.0:" + metricsPort,
							TimerType:     "histogram",
						},
						metricsPort,
					)),
				}, nil

			} else {

				return client.Options{
					HostPort:  targetHost,
					Namespace: namespace,
					ConnectionOptions: client.ConnectionOptions{
						TLS: &tls.Config{
							Certificates:       []tls.Certificate{cert},
							RootCAs:            serverCAPool,
							ServerName:         serverName,
							InsecureSkipVerify: insecureSkipVerify,
						},
					},
					Logger: NewTClientLogger(),

					// Set DataConverter to ensure that workflow inputs and results are
					// encrypted/decrypted as required.
					DataConverter: dataconverter.NewEncryptionDataConverter(
						converter.GetDefaultDataConverter(),
						dataconverter.DataConverterOptions{KeyID: os.Getenv("DATACONVERTER_ENCRYPTION_KEY_ID")},
					),
				}, nil
			}

		} else {
			// Self-hosted Temporal Server w/o TLS

			return client.Options{
				HostPort:  targetHost,
				Namespace: namespace,
				Logger:    NewTClientLogger(),
				DataConverter: dataconverter.NewEncryptionDataConverter(
					converter.GetDefaultDataConverter(),
					dataconverter.DataConverterOptions{KeyID: os.Getenv("DATACONVERTER_ENCRYPTION_KEY_ID")},
				),
			}, nil
		}

	} else {

		if useTLS {
			// Temporal Cloud

			return client.Options{
				HostPort:  targetHost,
				Namespace: namespace,
				ConnectionOptions: client.ConnectionOptions{
					TLS: &tls.Config{
						Certificates:       []tls.Certificate{cert},
						RootCAs:            serverCAPool,
						ServerName:         serverName,
						InsecureSkipVerify: insecureSkipVerify,
					},
				},
				Logger: NewTClientLogger(),
			}, nil

		} else {
			// Self-hosted Temporal Server w/o TLS

			return client.Options{
				HostPort:  targetHost,
				Namespace: namespace,
				Logger:    NewTClientLogger(),
			}, nil
		}

	}
}
