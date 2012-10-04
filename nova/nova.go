// Package nova provides types, methods and functions for interactions with the
// Nova OS API v2.
package nova

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/globocom/go-openstack/keystone"
	"io/ioutil"
	"net/http"
	"strings"
)

// Error returned by DisassociateNetwork when no network is found for the
// tenant.
var ErrNoNetwork = errors.New("Network not found: no network was found for this tenant.")

// NetworkDisassociator interface provides the method DisassociateNetwork, that
// is used to disassociate a network from a tenant.
type NetworkDisassociator interface {
	// Disassociates a network from the given tenant.
	DisassociateNetwork(tenantId string) error
}

type network struct {
	Id       string
	TenantId string `json:"project_id"`
}

// Client represents a client for the Nova OS API. It encapsulates a
// keystone.Client instance that provides the token and endpoints used by this
// client.
type Client struct {
	KeystoneClient *keystone.Client
}

func (c *Client) do(req *http.Request) ([]byte, int, error) {
	req.Header.Set("X-Auth-Token", c.KeystoneClient.Token)
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}

// DisassociateNetwork disassociates a network from the given tenant, returning
// an error in case of any failure.
func (c *Client) DisassociateNetwork(tenantId string) error {
	if c.KeystoneClient == nil {
		return errors.New("KeystoneClient is nil.")
	}
	endpoint := c.KeystoneClient.Endpoint("compute", "admin")
	req, err := http.NewRequest("GET", endpoint+"/os-networks", nil)
	if err != nil {
		return err
	}
	body, status, err := c.do(req)
	if err != nil {
		return errors.New("Failed to get the list of all networks: " + err.Error())
	}
	if status != http.StatusOK {
		return fmt.Errorf("Failed to get the list of all networks, status: %d.\nBody: %s.", status, body)
	}
	var result map[string][]network
	err = json.Unmarshal(body, &result)
	if err != nil {
		return errors.New("Failed to get the list of all networks, the server did not respond a valid JSON.")
	}
	var netId string
	for _, net := range result["networks"] {
		if net.TenantId == tenantId {
			netId = net.Id
			break
		}
	}
	if netId == "" {
		return ErrNoNetwork
	}
	reqBody := strings.NewReader(`{"disassociate":null}`)
	req, err = http.NewRequest("POST", endpoint+"/os-networks/"+netId+"/action", reqBody)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return err
	}
	body, status, err = c.do(req)
	if err != nil {
		return fmt.Errorf("Failed to disassociate the network %s from the tenant %s: %s", netId, tenantId, err)
	}
	if status != http.StatusAccepted {
		return fmt.Errorf("Failed to disassociate the network %s from the tenant %s, status: %d.\nBody: %s", netId, tenantId, status, body)
	}
	return nil
}
