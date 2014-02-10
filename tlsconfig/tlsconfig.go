package tlsconfig

import "crypto/tls"
import "crypto/x509"
import "io/ioutil"

func GetTLSConfig(clientCertPEMData, clientKeyPEMData []byte) (*tls.Config, error) {
	pemData, err := ioutil.ReadFile("orchard-certs.pem")
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(pemData)

	clientCert, err := tls.X509KeyPair(clientCertPEMData, clientKeyPEMData)
	if err != nil {
		return nil, err
	}

	config := new(tls.Config)
	config.RootCAs = certPool
	config.Certificates = []tls.Certificate{clientCert}
	config.BuildNameToCertificate()

	return config, nil
}
