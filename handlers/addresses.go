package handlers

import (
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/EnMasseProject/enmasse-rest/controller"
	"github.com/EnMasseProject/enmasse-rest/db"
	"github.com/EnMasseProject/enmasse-rest/models"
	"github.com/EnMasseProject/enmasse-rest/restapi/operations/addresses"
)

func NewErrorResponse(error *models.ErrorModel) *models.ErrorResponse {
	response := models.ErrorResponse{Errors: []*models.ErrorModel{error}}
	return &response
}

func NewErrorModel(status int32, title string, details string) *models.ErrorModel {
	return &models.ErrorModel{Status: &status, Details: &details, Title: &title}
}

func NewListErrorResponse(model *models.ErrorModel) middleware.Responder {
	return addresses.NewListAddressesDefault(int(*model.Status)).WithPayload(NewErrorResponse(model))
}

func NewCreateErrorResponse(model *models.ErrorModel) middleware.Responder {
	return addresses.NewCreateAddressDefault(int(*model.Status)).WithPayload(NewErrorResponse(model))
}

func NewPutErrorResponse(model *models.ErrorModel) middleware.Responder {
	return addresses.NewPutAddressesDefault(int(*model.Status)).WithPayload(NewErrorResponse(model))
}

func NewDeleteErrorResponse(model *models.ErrorModel) middleware.Responder {
	return addresses.NewDeleteAddressesDefault(int(*model.Status)).WithPayload(NewErrorResponse(model))
}

func PutAddressesHandler(addressDb db.AddressDB, ctrl controller.Controller, params addresses.PutAddressesParams) middleware.Responder {
	result, errModel := DeployAndSetConfig(addressDb, ctrl, &params.AddressConfigMap)
	if errModel != nil {
		return NewListErrorResponse(errModel)
	}
	return addresses.NewPutAddressesCreated().WithPayload(*result)
}

func DeployAndSetConfig(addressDb db.AddressDB, ctrl controller.Controller, config * models.AddressConfigMap) (* models.AddressConfigMap, *models.ErrorModel) {
	err := ctrl.DeployConfig(config)
	if err != nil {
		return nil, NewErrorModel(500, "Error deploying new configuration", err.Error())
	}

	result, err := addressDb.SetAddresses(config)
	if err != nil {
		return nil, NewErrorModel(500, "Error setting addresses", err.Error())
	}
	return result, nil
}

func DeleteAddressesHandler(addressDb db.AddressDB, ctrl controller.Controller, params addresses.DeleteAddressesParams) middleware.Responder {
	config, err := addressDb.GetAddresses()
	if err != nil {
		return NewDeleteErrorResponse(NewErrorModel(500, "Error retrieving addresses", err.Error()))
	}

	for _, address := range params.AddressList {
		delete(*config, address)
	}

	result, errModel := DeployAndSetConfig(addressDb, ctrl, config)
	if errModel != nil {
		return NewDeleteErrorResponse(errModel)
	}
	return addresses.NewDeleteAddressesOK().WithPayload(*result)
}

func ListAddressesHandler(addressDb db.AddressDB, params addresses.ListAddressesParams) middleware.Responder {
	config, err := addressDb.GetAddresses()
	if err != nil {
		return NewListErrorResponse(NewErrorModel(500, "Error retrieving addresses from DB", err.Error()))
	}

	return addresses.NewListAddressesOK().WithPayload(*config)
}

func CreateAddressHandler(addressDb db.AddressDB, ctrl controller.Controller, params addresses.CreateAddressParams) middleware.Responder {
	currentConfig, err := addressDb.GetAddresses()
	if err != nil {
		return NewCreateErrorResponse(NewErrorModel(500, "Error fetching addresses from DB", err.Error()))
	}

	for k, v := range params.AddressConfigMap {
		(*currentConfig)[k] = v
	}

	result, errModel := DeployAndSetConfig(addressDb, ctrl, currentConfig)
	if errModel != nil {
		return NewCreateErrorResponse(errModel)
	}

	return addresses.NewCreateAddressCreated().WithPayload(*result)
}
