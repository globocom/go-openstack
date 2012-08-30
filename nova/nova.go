package nova

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/timeredbull/openstack/keystone"
	"io/ioutil"
	"net/http"
	"strings"
)

type NetworkDisassociator interface {
	DisassociateNetwork(tenantId string) error
}

type network struct {
	Id       string
	TenantId string `json:"project_id"`
}

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
		return errors.New("Network not found: no network was found for this tenant.")
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
