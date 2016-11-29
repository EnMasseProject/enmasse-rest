package handlers

import (
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/EnMasseProject/enmasse-rest/models"
	"github.com/EnMasseProject/enmasse-rest/restapi/operations/addresses"
    "github.com/EnMasseProject/enmasse-rest/db"
    "github.com/EnMasseProject/enmasse-rest/controller"
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

func PutAddressesHandler(params addresses.PutAddressesParams) middleware.Responder {
    addressDb, err := db.GetAddressDB()
    if err != nil {
        return NewPutErrorResponse(NewErrorModel(500, "Error getting address DB", err.Error()))
    }

    result, errModel := DeployAndSetConfig(addressDb, params.AddressConfigMap)
    if errModel != nil {
        return NewListErrorResponse(errModel)
    }
    return addresses.NewPutAddressesCreated().WithPayload(result)
}

func DeployAndSetConfig(addressDb db.AddressDB, config models.AddressConfigMap) (models.AddressConfigMap, * models.ErrorModel) {
    err := controller.DeployConfig(config)
    if err != nil {
        return nil, NewErrorModel(500, "Error deploying new configuration", err.Error())
    }

    result, err := addressDb.SetAddresses(config)
    if err != nil {
        return nil, NewErrorModel(500, "Error setting addresses", err.Error())
    }
    return result, nil
}

func DeleteAddressesHandler(params addresses.DeleteAddressesParams) middleware.Responder {
    addressDb, err := db.GetAddressDB()
    if err != nil {
        return NewDeleteErrorResponse(NewErrorModel(500, "Error getting address DB", err.Error()))
    }

    config, err := addressDb.GetAddresses()
    if err != nil {
        return NewDeleteErrorResponse(NewErrorModel(500, "Error retrieving addresses", err.Error()))
    }

    for _, address := range params.AddressList {
        delete(config, address)
    }

    result, errModel := DeployAndSetConfig(addressDb, config)
    if errModel != nil {
        return NewDeleteErrorResponse(errModel)
    }
    return addresses.NewDeleteAddressesOK().WithPayload(result)
}

func ListAddressesHandler(params addresses.ListAddressesParams) middleware.Responder {
    addressDb, err := db.GetAddressDB()
    if err != nil {
        return NewListErrorResponse(NewErrorModel(500, "Error getting address DB", err.Error()))
    }

    config, err := addressDb.GetAddresses()
    if err != nil {
        return NewListErrorResponse(NewErrorModel(500, "Error retrieving addresses from DB", err.Error()))
    }

	return addresses.NewListAddressesOK().WithPayload(config)
}

func CreateAddressHandler(params addresses.CreateAddressParams) middleware.Responder {
    addressDb, err := db.GetAddressDB()
    if err != nil {
        return NewCreateErrorResponse(NewErrorModel(500, "Error getting address DB", err.Error()))
    }

    currentConfig, err := addressDb.GetAddresses()
    if err != nil {
        return NewCreateErrorResponse(NewErrorModel(500, "Error fetching addresses from DB", err.Error()))
    }

    for k, v := range params.AddressConfigMap {
        currentConfig[k] = v
    }

    result, errModel := DeployAndSetConfig(addressDb, currentConfig)
    if errModel != nil {
        return NewCreateErrorResponse(errModel)
    }

    return addresses.NewCreateAddressCreated().WithPayload(result)
}
