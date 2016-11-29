package addresses

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// CreateAddressHandlerFunc turns a function with the right signature into a create address handler
type CreateAddressHandlerFunc func(CreateAddressParams) middleware.Responder

// Handle executing the request and returning a response
func (fn CreateAddressHandlerFunc) Handle(params CreateAddressParams) middleware.Responder {
	return fn(params)
}

// CreateAddressHandler interface for that can handle valid create address params
type CreateAddressHandler interface {
	Handle(CreateAddressParams) middleware.Responder
}

// NewCreateAddress creates a new http.Handler for the create address operation
func NewCreateAddress(ctx *middleware.Context, handler CreateAddressHandler) *CreateAddress {
	return &CreateAddress{Context: ctx, Handler: handler}
}

/*CreateAddress swagger:route POST /v1/enmasse/addresses addresses createAddress

Create address config

*/
type CreateAddress struct {
	Context *middleware.Context
	Handler CreateAddressHandler
}

func (o *CreateAddress) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, _ := o.Context.RouteInfo(r)
	var Params = NewCreateAddressParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
