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
	DeployConfig(config *models.AddressConfigMap) error
    Close()
}

type storageController struct {
	conn electron.Connection
    sender electron.Sender
}

func (ctrl * storageController) Close() {
    ctrl.conn.Close(nil)
}

func GetController() (Controller, error) {
	var ctrl storageController
	host := os.Getenv("STORAGE_CONTROLLER_SERVICE_HOST")
	port := os.Getenv("STORAGE_CONTROLLER_SERVICE_PORT")
	if host == "" {
		host = "127.0.0.1" // os.Getenv("ADMIN_SERVICE_HOST")
	}
	if port == "" {
		port = os.Getenv("ADMIN_SERVICE_PORT_STORAGE_CONTROLLER")
	}

	if host == "" {
		return nil, errors.New("Neither ADMIN_SERVICE_HOST or STORAGE_CONTROLLER_SERVICE_HOST specified")
	}
	if port == "" {
		return nil, errors.New("Neither ADMIN_SERVICE_PORT_STORAGE_CONTROLLER or STORAGE_CONTROLLER_SERVICE_PORT specified")
	}
    addr := host + ":" + port
    conn, err := electron.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
    ctrl.conn = conn
    sender, err := ctrl.conn.Sender(electron.Target("address-config"))
    ctrl.sender = sender
	return &ctrl, err
}

func (ctrl *storageController) DeployConfig(config *models.AddressConfigMap) error {
	jbytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	payload := string(jbytes)

	outcome := ctrl.sender.SendSync(amqp.NewMessageWith(payload))
	return outcome.Error
}
