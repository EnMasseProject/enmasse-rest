package unittest

import (
    "errors"
    "testing"
    "fmt"
    loads "github.com/go-openapi/loads"
    "github.com/stretchr/testify/assert"
    httptransport "github.com/go-openapi/runtime/client"
    strfmt "github.com/go-openapi/strfmt"
    "github.com/EnMasseProject/enmasse-rest/client"
    addr "github.com/EnMasseProject/enmasse-rest/client/addresses"
    "github.com/EnMasseProject/enmasse-rest/models"
    "github.com/EnMasseProject/enmasse-rest/restapi"
    "github.com/EnMasseProject/enmasse-rest/db"
    "github.com/EnMasseProject/enmasse-rest/controller"
    "github.com/EnMasseProject/enmasse-rest/restapi/operations"
)

type fakeDb struct {
    config * models.AddressConfigMap
    err error
}

type fakeCtrl struct {
    config * models.AddressConfigMap
    err error
}

func (fdb * fakeDb) GetAddresses() (* models.AddressConfigMap, error) {
    if fdb.err != nil {
        return nil, fdb.err
    } else {
        return fdb.config, nil
    }
}

func (fdb * fakeDb) SetAddresses(config * models.AddressConfigMap) (*models.AddressConfigMap, error) {
    fdb.config = config
    if fdb.err != nil {
        return nil, fdb.err
    } else {
        return fdb.config, nil
    }
}

func (ctrl * fakeCtrl) Close() {
}

func (ctrl * fakeCtrl) DeployConfig(config * models.AddressConfigMap) error {
    if ctrl.err != nil {
        return ctrl.err
    } else {
        ctrl.config = config
        return nil
    }
}

func StartServer(adb db.AddressDB, ctrl controller.Controller) * restapi.Server {
    swaggerSpec, _ := loads.Analyzed(restapi.SwaggerJSON, "")
    api := operations.NewEnmasseRestAPI(swaggerSpec)
    server := restapi.NewServer(api)

    server.ConfigureAPI(adb, ctrl)

    server.Host = "127.0.0.1"
    server.EnabledListeners = []string{"http"}

    server.Listen()
    go server.Serve()
    return server
}

func getAddress(store bool, multicast bool, flavor string) models.AddressConfig {
    return models.AddressConfig {
        StoreAndForward: &store,
        Multicast: &multicast,
        Flavor: flavor,
    }
}

func GetClient(server * restapi.Server) * client.EnmasseRest {
    transport := httptransport.New(fmt.Sprintf("%s:%d", server.Host, server.Port), "/", []string{"http"})
    formats := strfmt.Default
    cli := client.New(transport, formats)
    return cli
}

func TestListHandler(t * testing.T) {
    assert := assert.New(t)
    fdb := fakeDb { nil, nil }
    ctrl := fakeCtrl { nil, nil }

    server := StartServer(&fdb, &ctrl)
    defer server.Shutdown()

    cli := GetClient(server)

    fdb.config = &models.AddressConfigMap {
        "myqueue": getAddress(true, false, "vanilla-queue"),
    }

    // Check that addresses are listed
    resp, err := cli.Addresses.ListAddresses(nil)
    assert.Nil(err)
    assert.NotNil(resp)
    assert.Equal(resp.Payload, *fdb.config)

    // Check that errors are propagated
    fdb.err = errors.New("Error1")
    resp, err = cli.Addresses.ListAddresses(nil)
    assert.Nil(resp)
    assert.NotNil(err)
}

func TestPutHandler(t * testing.T) {
    assert := assert.New(t)
    fdb := fakeDb { nil, nil }
    ctrl := fakeCtrl { nil, nil }

    server := StartServer(&fdb, &ctrl)
    defer server.Shutdown()

    cli := GetClient(server)

    config := models.AddressConfigMap {
        "myqueue": getAddress(true, false, "vanilla-queue"),
    }

    // Check that addresses can be replaced
    resp, err := cli.Addresses.PutAddresses(addr.NewPutAddressesParams().WithAddressConfigMap(config))
    assert.Nil(err)
    assert.NotNil(resp)
    assert.Equal(resp.Payload, config)
    assert.Equal(*fdb.config, config)
    assert.Equal(*ctrl.config, config)

    // Ensure that errors are propagated
    fdb.err = errors.New("Error1")
    resp, err = cli.Addresses.PutAddresses(addr.NewPutAddressesParams().WithAddressConfigMap(config))
    assert.Nil(resp)
    assert.NotNil(err)

    // Ensure that previous config is still valid if deployment fails
    fdb.err = nil
    fdb.config = nil
    ctrl.err = errors.New("Error2")
    resp, err = cli.Addresses.PutAddresses(addr.NewPutAddressesParams().WithAddressConfigMap(config))
    assert.Nil(resp)
    assert.NotNil(err)
    assert.Nil(fdb.config)
}

func TestCreateHandler(t * testing.T) {
    assert := assert.New(t)
    fdb := fakeDb { nil, nil }
    ctrl := fakeCtrl { nil, nil }

    server := StartServer(&fdb, &ctrl)
    defer server.Shutdown()

    cli := GetClient(server)

    config := models.AddressConfigMap {
        "myqueue": getAddress(true, false, "vanilla-queue"),
    }

    fdb.config = &config

    newConfig := models.AddressConfigMap {
        "mytopic": getAddress(true, true, "vanilla-topic"),
    }

    // Check that addresses can be appended
    resp, err := cli.Addresses.CreateAddress(addr.NewCreateAddressParams().WithAddressConfigMap(newConfig))
    assert.Nil(err)
    assert.NotNil(resp)

    expectedConfig := models.AddressConfigMap {
        "myqueue": getAddress(true, false, "vanilla-queue"),
        "mytopic": getAddress(true, true, "vanilla-topic"),
    }
    assert.Equal(resp.Payload, expectedConfig)
    assert.Equal(*fdb.config, expectedConfig)
    assert.Equal(*ctrl.config, expectedConfig)

    // Test if create errors are handled correctly
    fdb.err = errors.New("Error1")
    resp, err = cli.Addresses.CreateAddress(addr.NewCreateAddressParams().WithAddressConfigMap(newConfig))
    assert.Nil(resp)
    assert.NotNil(err)

    // Verify that old address is kept on errors
    fdb.err = nil
    fdb.config = &config
    ctrl.err = errors.New("Error2")
    resp, err = cli.Addresses.CreateAddress(addr.NewCreateAddressParams().WithAddressConfigMap(newConfig))
    assert.Nil(resp)
    assert.NotNil(err)
    assert.Equal(*fdb.config, config)
}

func TestDeleteHandler(t * testing.T) {
    assert := assert.New(t)

    config := models.AddressConfigMap {
        "myqueue": getAddress(true, false, "vanilla-queue"),
        "mytopic": getAddress(true, true, "vanilla-topic"),
    }
    fdb := fakeDb { &config, nil }
    ctrl := fakeCtrl { &config, nil }

    server := StartServer(&fdb, &ctrl)
    defer server.Shutdown()

    cli := GetClient(server)

    fdb.config = &config

    toDelete := models.AddressList {
        "myqueue",
    }

    // Test that addresses can be deleted
    resp, err := cli.Addresses.DeleteAddresses(addr.NewDeleteAddressesParams().WithAddressList(toDelete))
    assert.Nil(err)
    assert.NotNil(resp)

    expectedConfig := models.AddressConfigMap {
        "mytopic": getAddress(true, true, "vanilla-topic"),
    }
    assert.Equal(resp.Payload, expectedConfig)
    assert.Equal(*fdb.config, expectedConfig)
    assert.Equal(*ctrl.config, expectedConfig)

    // Test errors when deleting address
    fdb.err = errors.New("Error1")
    resp, err = cli.Addresses.DeleteAddresses(addr.NewDeleteAddressesParams().WithAddressList(toDelete))
    assert.Nil(resp)
    assert.NotNil(err)

    // Verify that old address is kept on errors
    fdb.err = nil
    fdb.config = &config
    ctrl.err = errors.New("Error2")
    resp, err = cli.Addresses.DeleteAddresses(addr.NewDeleteAddressesParams().WithAddressList(toDelete))
    assert.Nil(resp)
    assert.NotNil(err)
    assert.Equal(*fdb.config, config)
}
