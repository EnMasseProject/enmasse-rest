package handlers

import (
	"encoding/json"
	middleware "github.com/go-openapi/runtime/middleware"

	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"

	"fmt"
	"github.com/EnMasseProject/enmasse-rest/models"
	"github.com/EnMasseProject/enmasse-rest/restapi/operations/addresses"
	"os"
	"qpid.apache.org/amqp"
	"qpid.apache.org/electron"
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

func ListAddressesHandler(params addresses.ListAddressesParams) middleware.Responder {
	configMaps, errorModel := GetConfigMaps()
	if errorModel != nil {
		return NewListErrorResponse(errorModel)
	}
	var config models.AddressConfigMap
	configMap, err := configMaps.Get("addressdb")
	if err == nil {
		jstr := configMap.Data["json"]
		if err := json.Unmarshal([]byte(jstr), &config); err != nil {
			return NewListErrorResponse(NewErrorModel(500, "Error reading config", err.Error()))
		}
	}

	return addresses.NewListAddressesOK().WithPayload(config)
}

func GetControllerAddress() string {
	host := os.Getenv("STORAGE_CONTROLLER_SERVICE_HOST")
	port := os.Getenv("STORAGE_CONTROLLER_SERVICE_PORT")
	return host + ":" + port
}

func DeployConfig(config string) error {
	addr := GetControllerAddress()
	c, err := electron.Dial("tcp", addr)
	if err != nil {
        return err
	}
	defer c.Close(nil)
	s, err := c.Sender(electron.Target("address-config"))
	if err != nil {
        return err
	}
    outcome := s.SendSync(amqp.NewMessageWith(config))
    return outcome.Error
}

func CreateAddressHandler(params addresses.CreateAddressParams) middleware.Responder {
	configMaps, errorModel := GetConfigMaps()
	if errorModel != nil {
		return NewCreateErrorResponse(errorModel)
	}

    currentConfig := make(models.AddressConfigMap)
	configMap, err := configMaps.Get("addressdb")
    newConfig := err != nil
    if err == nil {
        if data, ok := configMap.Data["json"]; ok {
            fmt.Printf("Data was set, decoding\n")
            if err := json.Unmarshal([]byte(data), &currentConfig); err != nil {
                return NewCreateErrorResponse(NewErrorModel(500, "Error decoding existing configuration", err.Error()))
            }
        } else {
            return NewCreateErrorResponse(NewErrorModel(500, "Error retrieving config", err.Error()))
        }
    }

    for k, v := range params.AddressConfigMap {
        currentConfig[k] = v
    }

	jbytes , err := json.Marshal(currentConfig)
	if err != nil {
		return NewCreateErrorResponse(NewErrorModel(500, "Error serializing address config", err.Error()))
	}

    jstr := string(jbytes)

    fmt.Printf("Deploying new address config: %s\n", jstr)
    err = DeployConfig(jstr)
    if err != nil {
		return NewCreateErrorResponse(NewErrorModel(500, "Error deploying address config", err.Error()))
    }

    if newConfig {
        configMap = new(v1.ConfigMap)
        configMap.Name = "addressdb"
        configMap.Namespace, _ = GetNamespace()
        configMap.Data = make(map[string]string)
        configMap.Data["json"] = jstr
        _, err = configMaps.Create(configMap)
		if err != nil {
			return NewCreateErrorResponse(NewErrorModel(500, "Unable to create configmap", err.Error()))
		}
    } else {
        configMap.Data = make(map[string]string)
        configMap.Data["json"] = jstr
        _, err = configMaps.Update(configMap)
        if err != nil {
            return NewCreateErrorResponse(NewErrorModel(500, "Unable to update configmap", err.Error()))
        }
    }
	return addresses.NewCreateAddressCreated().WithPayload(currentConfig)
}
