package handlers

import (
    middleware "github.com/go-openapi/runtime/middleware"
    "encoding/json"

    "io/ioutil"
    "fmt"
    "k8s.io/client-go/1.4/kubernetes"
    "k8s.io/client-go/1.4/rest"

    "enmasse-rest-api/restapi/operations/addresses"
    "enmasse-rest-api/models"
)

const NS_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

func getClient() (*kubernetes.Clientset, error) {
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

func GetAddressesHandler(params addresses.GetAddressesParams) middleware.Responder {
    client, err := getClient()
    if err != nil {
        fmt.Printf("Unable to create client")
        return addresses.NewGetAddressesDefault(500)
    }
	namespace, err := GetNamespace()
    if err != nil {
        fmt.Printf("Unable to find namespace")
        return addresses.NewGetAddressesDefault(500)
    }
    configMap, err := client.Core().ConfigMaps(namespace).Get("maas")
    if err != nil {
        fmt.Printf("Unable to get configmap")
        return addresses.NewGetAddressesDefault(500)
    }

    jstr := configMap.Data["json"]
    var config models.AddressConfig
    if err := json.Unmarshal([]byte(jstr), &config); err != nil {
        fmt.Printf("Error converting map to json. Have: %s", jstr)
        return addresses.NewGetAddressesDefault(500)
    }
    return addresses.NewGetAddressesOK().WithPayload(config)
}

func PutAddressesHandler(params addresses.PutAddressesParams) middleware.Responder {
    jstr, err := json.Marshal(params.AddressConfig)
    if err != nil {
        fmt.Printf("Error serializing address config")
        return addresses.NewPutAddressesDefault(500)
    }
    client, err := getClient()
    if err != nil {
        fmt.Printf("Unable to create client")
        return addresses.NewPutAddressesDefault(500)
    }
	namespace, err := GetNamespace()
    if err != nil {
        fmt.Printf("Unable to find namespace")
        return addresses.NewPutAddressesDefault(500)
    }
    configMap, err := client.Core().ConfigMaps(namespace).Get("maas")
    if err != nil {
        fmt.Printf("Unable to get configmap")
        return addresses.NewPutAddressesDefault(500)
    }

    configMap.Data["json"] = string(jstr)
    _, err = client.Core().ConfigMaps(namespace).Update(configMap)
    if err != nil {
        fmt.Printf("Unable to update configmap")
        return addresses.NewPutAddressesDefault(500)
    }
    return addresses.NewPutAddressesCreated().WithPayload(params.AddressConfig)
}
