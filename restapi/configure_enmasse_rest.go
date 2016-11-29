package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/EnMasseProject/enmasse-rest/controller"
	"github.com/EnMasseProject/enmasse-rest/db"
	"github.com/EnMasseProject/enmasse-rest/handlers"
	"github.com/EnMasseProject/enmasse-rest/restapi/operations"
	"github.com/EnMasseProject/enmasse-rest/restapi/operations/addresses"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target .. --name  --spec ../api/swagger.yaml

func configureFlags(api *operations.EnmasseRestAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.EnmasseRestAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// s.api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	ctrl, err := controller.GetController()
	if err != nil {
		panic(err)
	}

	adb, err := db.GetAddressDB()
	if err != nil {
		panic(err)
	}

	api.AddressesListAddressesHandler = addresses.ListAddressesHandlerFunc(func(p addresses.ListAddressesParams) middleware.Responder {
		return handlers.ListAddressesHandler(adb, p)
	})
	api.AddressesCreateAddressHandler = addresses.CreateAddressHandlerFunc(func(p addresses.CreateAddressParams) middleware.Responder {
		return handlers.CreateAddressHandler(adb, ctrl, p)
	})
	api.AddressesPutAddressesHandler = addresses.PutAddressesHandlerFunc(func(p addresses.PutAddressesParams) middleware.Responder {
		return handlers.PutAddressesHandler(adb, ctrl, p)
	})
	api.AddressesDeleteAddressesHandler = addresses.DeleteAddressesHandlerFunc(func(p addresses.DeleteAddressesParams) middleware.Responder {
		return handlers.DeleteAddressesHandler(adb, ctrl, p)
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
