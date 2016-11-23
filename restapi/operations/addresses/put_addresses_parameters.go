package addresses

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/enmasseproject/enmasse-rest/models"
)

// NewPutAddressesParams creates a new PutAddressesParams object
// with the default values initialized.
func NewPutAddressesParams() PutAddressesParams {
	var ()
	return PutAddressesParams{}
}

// PutAddressesParams contains all the bound params for the put addresses operation
// typically these are obtained from a http.Request
//
// swagger:parameters putAddresses
type PutAddressesParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request

	/*AddressConfig to set
	  Required: true
	  In: body
	*/
	AddressConfig models.AddressConfig
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *PutAddressesParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error
	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.AddressConfig
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("addressConfig", "body"))
			} else {
				res = append(res, errors.NewParseError("addressConfig", "body", "", err))
			}

		} else {
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.AddressConfig = body
			}
		}

	} else {
		res = append(res, errors.Required("addressConfig", "body"))
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
