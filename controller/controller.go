package controller

import (
	"encoding/json"

	"github.com/EnMasseProject/enmasse-rest/models"
	"os"
	"qpid.apache.org/amqp"
	"qpid.apache.org/electron"
)

func getControllerAddress() string {
	host := os.Getenv("STORAGE_CONTROLLER_SERVICE_HOST")
	port := os.Getenv("STORAGE_CONTROLLER_SERVICE_PORT")
	return host + ":" + port
}

func DeployConfig(config models.AddressConfigMap) error {
	addr := getControllerAddress()
	c, err := electron.Dial("tcp", addr)
	if err != nil {
        return err
	}
	defer c.Close(nil)
	s, err := c.Sender(electron.Target("address-config"))
	if err != nil {
        return err
	}

	jbytes , err := json.Marshal(config)
	if err != nil {
		return err
	}

    payload := string(jbytes)

    outcome := s.SendSync(amqp.NewMessageWith(payload))
    return outcome.Error
}
