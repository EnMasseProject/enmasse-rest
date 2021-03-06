package client

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/EnMasseProject/enmasse-rest/client/addresses"
)

// Default enmasse rest HTTP client.
var Default = NewHTTPClient(nil)

// NewHTTPClient creates a new enmasse rest HTTP client.
func NewHTTPClient(formats strfmt.Registry) *EnmasseRest {
	if formats == nil {
		formats = strfmt.Default
	}
	transport := httptransport.New("localhost", "/", []string{"http"})
	return New(transport, formats)
}

// New creates a new enmasse rest client
func New(transport runtime.ClientTransport, formats strfmt.Registry) *EnmasseRest {
	cli := new(EnmasseRest)
	cli.Transport = transport

	cli.Addresses = addresses.New(transport, formats)

	return cli
}

// EnmasseRest is a client for enmasse rest
type EnmasseRest struct {
	Addresses *addresses.Client

	Transport runtime.ClientTransport
}

// SetTransport changes the transport on the client and all its subresources
func (c *EnmasseRest) SetTransport(transport runtime.ClientTransport) {
	c.Transport = transport

	c.Addresses.SetTransport(transport)

}
