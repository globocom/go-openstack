package keystone

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Token   string
	authUrl string
}

type Tenant struct {
	Id          string
	Name        string
	Description string
}

type User struct {
	Id    string
	Name  string
	Email string
}

type Ec2 struct {
	Access string
	Secret string
}

func NewClient(username, password, tenantName, authUrl string) (*Client, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"auth": {"passwordCredentials": {"username": "%s", "password":"%s"}, "tenantName": "%s"}}`, username, password, tenantName))
	response, err := http.Post(authUrl+"/tokens", "application/json", b)
	defer response.Body.Close()
	result, _ := ioutil.ReadAll(response.Body)
	var data map[string]map[string]interface{}
	err = json.Unmarshal(result, &data)
	if err != nil {
		return nil, err
	}
	if response.StatusCode > 399 {
		return nil, errors.New(data["error"]["title"].(string))
	}
	token := data["access"]["token"].(map[string]interface{})["id"].(string)
	return &Client{Token: token, authUrl: authUrl}, nil
}

func (c *Client) do(method, urlStr string, body io.Reader) (*http.Response, error) {
	request, _ := http.NewRequest(method, urlStr, body)
	request.Header.Set("X-Auth-Token", c.Token)
	request.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	return httpClient.Do(request)
}

func (c *Client) NewTenant(name, description string, enabled bool) (*Tenant, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"tenant": {"name": "%s", "description": "%s", "enabled": %t}}`, name, description, enabled))
	response, _ := c.do("POST", c.authUrl+"/tenants", b)
	defer response.Body.Close()
	result, _ := ioutil.ReadAll(response.Body)
	var data map[string]map[string]interface{}
	_ = json.Unmarshal(result, &data)
	tenant := Tenant{
		Id:          data["tenant"]["id"].(string),
		Name:        data["tenant"]["name"].(string),
		Description: data["tenant"]["description"].(string),
	}
	return &tenant, nil
}

func (c *Client) NewUser(name, password, email, tenantId string, enabled bool) (*User, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"user": {"name": "%s", "password": "%s", "tenantId": "%s", "email": "%s", "enabled": %t}}`, name, password, tenantId, email, enabled))
	response, _ := c.do("POST", c.authUrl+"/users", b)
	defer response.Body.Close()
	result, _ := ioutil.ReadAll(response.Body)
	var data map[string]map[string]interface{}
	_ = json.Unmarshal(result, &data)
	user := User{
		Id:    data["user"]["id"].(string),
		Name:  data["user"]["name"].(string),
		Email: data["user"]["email"].(string),
	}
	return &user, nil
}

func (c *Client) NewEc2(userId, tenantId string) (*Ec2, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"tenant_id": "%s"}`, tenantId))
	response, _ := c.do("POST", c.authUrl+"/users/"+userId+"/credentials/OS-EC2", b)
	defer response.Body.Close()
	result, _ := ioutil.ReadAll(response.Body)
	var data map[string]map[string]interface{}
	_ = json.Unmarshal(result, &data)
	ec2 := Ec2{
		Access: data["credential"]["access"].(string),
		Secret: data["credential"]["secret"].(string),
	}
	return &ec2, nil
}

func (c *Client) RemoveEc2(userId, access string) error {
	return c.delete(c.authUrl+"/users/"+userId+"/credentials/OS-EC2/"+access)
}

func (c *Client) RemoveUser(userId string) error {
	return c.delete(c.authUrl+"/users/"+userId)
}

func (c *Client) RemoveTenant(tenantId string) error {
	return c.delete(c.authUrl+"/tenants/"+tenantId)
}

func (c *Client) delete(url string) error {
	if resp, err := c.do("DELETE", url, nil); err != nil {
		return err
	} else if resp.StatusCode > 299 {
		return errorFromResponse(resp)
	}
	return nil
}

func errorFromResponse(response *http.Response) error {
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return errors.New(string(b))
}
