package handlers

import (
    middleware "github.com/go-openapi/runtime/middleware"
    "encoding/json"

    "io/ioutil"
    "k8s.io/client-go/1.4/kubernetes"
    v1 "k8s.io/client-go/1.4/pkg/api/v1"
    v1core "k8s.io/client-go/1.4/kubernetes/typed/core/v1"
    "k8s.io/client-go/1.4/rest"

    "github.com/EnMasseProject/enmasse-rest/restapi/operations/addresses"
    "github.com/EnMasseProject/enmasse-rest/models"
)

const NS_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

func GetClient() (*kubernetes.Clientset, error) {
    // creates the in-cluster config
    config, err := rest.InClusterConfig()
    if err != nil {
        return nil, err
    }
    // creates the clientset
    return kubernetes.NewForConfig(config)
}

func GetNamespace() (string, error) {
	ns, err := ioutil.ReadFile(NS_PATH)
    if ns != nil && err == nil {
        return string(ns), err
    }
    return "", err
}

func NewErrorResponse(error * models.ErrorModel) *models.ErrorResponse {
	response := models.ErrorResponse{Errors: []*models.ErrorModel{error}}
	return &response
}

func NewErrorModel(status int32, title string, details string) * models.ErrorModel {
	return &models.ErrorModel{Status: &status, Details: &details, Title: &title}
}

func NewGetErrorResponse(model * models.ErrorModel) middleware.Responder {
    return addresses.NewGetAddressesDefault(int(*model.Status)).WithPayload(NewErrorResponse(model))
}

func NewPutErrorResponse(model * models.ErrorModel) middleware.Responder {
    return addresses.NewPutAddressesDefault(int(*model.Status)).WithPayload(NewErrorResponse(model))
}

func GetConfigMaps() (v1core.ConfigMapInterface, *models.ErrorModel) {
    client, err := GetClient()
    if err != nil {
        return nil, NewErrorModel(500, "Unable to create client", err.Error())
    }
	namespace, err := GetNamespace()
    if err != nil {
        return nil, NewErrorModel(500, "Unable to find namespace", err.Error())
    }
    return client.Core().ConfigMaps(namespace), nil
}

func GetAddressesHandler(params addresses.GetAddressesParams) middleware.Responder {
    configMaps, errorModel := GetConfigMaps()
    if errorModel != nil {
        return NewGetErrorResponse(errorModel)
    }
    var config models.AddressConfig
    configMap, err := configMaps.Get("maas")
    if err == nil {
        jstr := configMap.Data["json"]
        if err := json.Unmarshal([]byte(jstr), &config); err != nil {
            return NewGetErrorResponse(NewErrorModel(500, "Error reading config", err.Error()))
        }
    }
    return addresses.NewGetAddressesOK().WithPayload(config)
}

func PutAddressesHandler(params addresses.PutAddressesParams) middleware.Responder {
    jstr, err := json.Marshal(params.AddressConfig)
    if err != nil {
        return NewPutErrorResponse(NewErrorModel(500, "Error serializing address config", err.Error()))
    }
    configMaps, errorModel := GetConfigMaps()
    if errorModel != nil {
        return NewGetErrorResponse(errorModel)
    }

    configMap, err := configMaps.Get("maas")
    if err == nil {
        configMap.Data["json"] = string(jstr)
        _, err = configMaps.Update(configMap)
        if err != nil {
            return NewPutErrorResponse(NewErrorModel(500, "Unable to update configmap", err.Error()))
        }
    } else {
        configMap := new(v1.ConfigMap)
        configMap.Name = "maas"
        configMap.Data = make(map[string]string)
        configMap.Data["json"] = string(jstr)
        _, err = configMaps.Create(configMap)
        if err != nil {
            return NewPutErrorResponse(NewErrorModel(500, "Unable to create configmap", err.Error()))
        }
    }
    return addresses.NewPutAddressesCreated().WithPayload(params.AddressConfig)
}
