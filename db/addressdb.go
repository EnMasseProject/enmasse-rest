package db

import (
	"encoding/json"
	"github.com/EnMasseProject/enmasse-rest/models"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

const NS_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

type AddressDB interface {
	SetAddresses(config * models.AddressConfigMap) (* models.AddressConfigMap, error)
	GetAddresses() (* models.AddressConfigMap, error)
}

type configMapDB struct {
	configMaps v1core.ConfigMapInterface
	configMap  *v1.ConfigMap
}

func GetAddressDB() (AddressDB, error) {
	var adb configMapDB
	var err error

	adb.configMaps, err = getConfigMaps()
	if err != nil {
		return nil, err
	}
	adb.configMap, err = adb.configMaps.Get("addressdb")
    if err != nil {
        adb.configMap = nil
    }
	return &adb, nil
}

func (adb *configMapDB) SetAddresses(config * models.AddressConfigMap) (* models.AddressConfigMap, error) {
	jbytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	jstr := string(jbytes)

	if adb.configMap == nil {
		adb.configMap = new(v1.ConfigMap)
		adb.configMap.Name = "addressdb"
		adb.configMap.Namespace, _ = getNamespace()
		adb.configMap.Data = make(map[string]string)
		adb.configMap.Data["json"] = jstr
		_, err = adb.configMaps.Create(adb.configMap)
		if err != nil {
			return nil, err
		}
	} else {
		adb.configMap.Data = make(map[string]string)
		adb.configMap.Data["json"] = jstr
		_, err = adb.configMaps.Update(adb.configMap)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func (adb *configMapDB) GetAddresses() (* models.AddressConfigMap, error) {
	var config = new(models.AddressConfigMap)
	if adb.configMap != nil {
		jstr := adb.configMap.Data["json"]
		if err := json.Unmarshal([]byte(jstr), &config); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func getClient() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	return kubernetes.NewForConfig(config)
}

func getNamespace() (string, error) {
	ns, err := ioutil.ReadFile(NS_PATH)
	if ns != nil && err == nil {
		return string(ns), err
	}
	return "", err
}

func getConfigMaps() (v1core.ConfigMapInterface, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	namespace, err := getNamespace()
	if err != nil {
		return nil, err
	}
	return client.Core().ConfigMaps(namespace), nil
}
