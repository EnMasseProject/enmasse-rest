package controller

import (
	"encoding/json"
	"errors"

	"github.com/EnMasseProject/enmasse-rest/models"
	"os"
	"qpid.apache.org/amqp"
	"qpid.apache.org/electron"
)

type Controller interface {
	DeployConfig(config * models.AddressConfigMap) error
}

type storageController struct {
	addr string
}

func GetController() (Controller, error) {
	var ctrl storageController
	host := os.Getenv("STORAGE_CONTROLLER_SERVICE_HOST")
	port := os.Getenv("STORAGE_CONTROLLER_SERVICE_PORT")
	if host == "" {
		return nil, errors.New("STORAGE_CONTROLLER_SERVICE_HOST not specified")
	}
	if port == "" {
		return nil, errors.New("STORAGE_CONTROLLER_SERVICE_PORT not specified")
	}
	ctrl.addr = host + ":" + port
	return &ctrl, nil
}

func (ctrl *storageController) DeployConfig(config * models.AddressConfigMap) error {
	c, err := electron.Dial("tcp", ctrl.addr)
	if err != nil {
		return err
	}
	defer c.Close(nil)
	s, err := c.Sender(electron.Target("address-config"))
	if err != nil {
		return err
	}

	jbytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	payload := string(jbytes)

	outcome := s.SendSync(amqp.NewMessageWith(payload))
	return outcome.Error
}
