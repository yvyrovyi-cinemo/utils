package kafkalib

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func BuildTLSConfig(config Config) (*tls.Config, error) {

	clientCert, err := os.ReadFile(config.ClientCertFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading client cert file: %w", err)
	}

	clientKey, err := os.ReadFile(config.ClientKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading client key file: %w", err)
	}

	caCert, err := os.ReadFile(config.CaFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading CA cert file: %w", err)
	}

	certificate, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, fmt.Errorf("creating key pair: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("creating certificate: %w", err)
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}, nil
}
