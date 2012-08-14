package keystone

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func NewClient(username, password, tenantName, authUrl string) (string, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"auth": {"passwordCredentials": {"username": "%s", "password":"%s"}, "tenantName": "%s"}}`, username, password, tenantName))
	response, err := http.Post(authUrl+"/tokens", "application/json", b)
	defer response.Body.Close()
	result, _ := ioutil.ReadAll(response.Body)
	var data map[string]map[string]interface{}
	err = json.Unmarshal(result, &data)
	if err != nil {
		return "", err
	}
	if response.StatusCode > 399 {
		return "", errors.New(data["error"]["title"].(string))
	}
	return "", err
}
