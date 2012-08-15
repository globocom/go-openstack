package keystone

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	return &Client{Token: token, authUrl: authUrl}, err
}

func (c *Client) NewTenant(name, description string, enabled bool) (*Tenant, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"tenant": {"name": "%s", "description": "%s", "enabled": %t}}`, name, description, enabled))
	request, err := http.NewRequest("POST", c.authUrl+"/tenants", b)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Auth-Token", c.Token)
	request.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	defer response.Body.Close()
	result, _ := ioutil.ReadAll(response.Body)
	var data map[string]map[string]interface{}
	err = json.Unmarshal(result, &data)
	if err != nil {
		return nil, err
	}
	tenant := Tenant{
		Id:          data["tenant"]["id"].(string),
		Name:        data["tenant"]["name"].(string),
		Description: data["tenant"]["description"].(string),
	}
	return &tenant, nil
}

func (c *Client) NewUser(name, password, email, tenantId string, enabled bool) (*User, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"user": {"name": "%s", "password": "%s", "tenantId": "%s", "email": "%s", "enabled": %t}}`, name, password, tenantId, email, enabled))
	request, err := http.NewRequest("POST", c.authUrl+"/users", b)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Auth-Token", c.Token)
	request.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	defer response.Body.Close()
	result, _ := ioutil.ReadAll(response.Body)
	var data map[string]map[string]interface{}
	err = json.Unmarshal(result, &data)
	if err != nil {
		return nil, err
	}
	user := User{
		Id:    data["user"]["id"].(string),
		Name:  data["user"]["name"].(string),
		Email: data["user"]["email"].(string),
	}
	return &user, nil
}
