package config

import (
	"fmt"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/resonatecoop/user-api-client/client"
)

var (
	basepath = ""
	schemes  = []string{}
)

// NewAPIClient
func NewAPIClient(config UserAPIConfig) *apiclient.ResonateServiceDocumentationUser {
	if config.Hostname == "" {
		panic("user api hostname not set")
	}

	httpClient, err := httptransport.TLSClient(httptransport.TLSClientOptions{
		InsecureSkipVerify: config.InsecureSkipVerify,
	})

	if err != nil {
		panic(err)
	}

	hostname := config.Hostname

	if config.Port != "" {
		hostname = fmt.Sprintf("%s:%s", config.Hostname, config.Port)
	}

	transport := httptransport.NewWithClient(hostname, basepath, schemes, httpClient)

	client := apiclient.New(transport, strfmt.Default)

	return client
}
