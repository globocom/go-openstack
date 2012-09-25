// Package keystone provides types, methods and functions for interactions with
// Keystone API v2.0. It allows a developer to create and delete tenants, users
// and EC2 credentials (access key and secret key).
//
// This client does not store password for users in any of its types.
package keystone

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// ServiceCatalog represents a service catalog. Each service has a name and a
// type, and a collection of endpoints (one per region).
//
// Example of ServiceCatalog instance (for nova):
//
//     ServiceCatalog{
//         Name:      "Compute Service",
//         Type:      "compute",
//         Endpoints: []map[string]string{
//             map[string]string{
//                 "region":      "RegionOne",
//                 "adminURL":    "http://mynova.com:8774/v2/tenant-id",
//                 "publicURL":   "http://mynova.com:8774/v2/tenant-id",
//                 "internalURL": "http://mynova.com:8774/v2/tenant-id",
//             },
//         },
//     }
type ServiceCatalog struct {
	Endpoints []map[string]string
	Type      string
	Name      string
}

// Client represents a keystone connection client. It stores the authenticatin
// token and a lista of service catalogs (see ServiceCatalog type).
type Client struct {
	// Token is the authentication token that should be used in requests to
	// OpenStack services API.
	Token string

	// Catalogs is a slice of ServiceCatalog all catalogs of services available
	// for the authenticated user (see NewClient function for authentication
	// details).
	Catalogs []ServiceCatalog

	authUrl string
}

// Tenant represents a keystone tenant.
type Tenant struct {
	Id          string
	Name        string
	Description string
}

// User represents a keystone user. Please notice that it does not store the
// user password.
type User struct {
	Id    string
	Name  string
	Email string
}

// AddRoleToUser associates a role with a user and tenant
// Returns an error in case of failure
func (c *Client) AddRoleToUser(user_id, tenant_id, role_id string) error {
	roleUrl := fmt.Sprintf("/tenants/%s/users/%s/roles/OS-KSADM/%s", tenant_id, user_id, role_id)
	_, err := c.do("PUT", c.authUrl+roleUrl, nil)
	return err
}

// Ec2 represents a EC2 credential pair, containing an access key and a secret
// key.
type Ec2 struct {
	Access string
	Secret string
}

// NewClient returns a new instance of the client, authenticating in the
// provided authUrl.
//
// For authentication, it uses the parameters username, password and tenantName
// to issue a request to the authUrl. The new generated token is stored in the
// Client instance, as is the service catalog.
func NewClient(username, password, tenantName, authUrl string) (*Client, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"auth": {"passwordCredentials": {"username": "%s", "password":"%s"}, "tenantName": "%s"}}`, username, password, tenantName))
	response, err := http.Post(authUrl+"/tokens", "application/json", b)
	if err != nil {
		panic(err)
	}
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
	client := Client{Token: token, authUrl: authUrl}
	catalogs := data["access"]["serviceCatalog"].([]interface{})
	for _, c := range catalogs {
		catalog := c.(map[string]interface{})
		serviceCatalog := ServiceCatalog{
			Name: catalog["name"].(string),
			Type: catalog["type"].(string),
		}
		for _, e := range catalog["endpoints"].([]interface{}) {
			endpoint := map[string]string{}
			for k, v := range e.(map[string]interface{}) {
				endpoint[k] = v.(string)
			}
			serviceCatalog.Endpoints = append(serviceCatalog.Endpoints, endpoint)
		}
		client.Catalogs = append(client.Catalogs, serviceCatalog)
	}
	return &client, nil
}

// Endpoint returns the endpoint string for the given service and type of URL.
//
// The endpoint is get from the service catalog. If the given service or URL
// type is not present in the catalog, Endpoint returns an empty string.
//
// Examples of use:
//
//     var endpoint string
//     endpoint = client.Endpoint("compute", "adminURL")
//     endpoint = client.Endpoint("compute", "admin") // note that you can omit "URL" and it still works
//     endpoint = client.Endpoint("unknownservice", "adminURL") // returns ""
//     endpoint = client.Endpoint("compute", "unknownURL") // returns ""
func (c *Client) Endpoint(service, which string) string {
	var (
		endpoint string
		catalog  ServiceCatalog
	)
	for _, catalog = range c.Catalogs {
		if catalog.Type == service {
			break
		}
	}
	if catalog.Type == service {
		if !strings.Contains(which, "URL") {
			which += "URL"
		}
		endpoint = catalog.Endpoints[0][which] // TODO(fsouza): choose region
	}
	return endpoint
}

func (c *Client) do(method, urlStr string, body io.Reader) (*http.Response, error) {
	request, _ := http.NewRequest(method, urlStr, body)
	request.Header.Set("X-Auth-Token", c.Token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	httpClient := &http.Client{}
	return httpClient.Do(request)
}

// NewTenant creates a new tenant using the given name and description. The
// third parameter is a flag that indicates if the tenant should be enabled or
// not.
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

// NewUser create a new user using the given name, password and email.
//
// The last parameter is a flag that indicates if the user should be enabled or
// not.
//
// Besides creating the user, this method will also create a new role for the
// user using the given tenant and role. For example, if you have a "member"
// role with ID `834452bcb9f94178aaa4167cff1034df`, and a tenant with ID
// `2cc6842387314f868c9be75684c64530`, the following call will create a user
// called "gopher" and add it as a member of the tenant:
//
//     var role string = "834452bcb9f94178aaa4167cff1034df"
//     var tenant string = "2cc6842387314f868c9be75684c64530"
//     client.NewUser("gopher", "secret", "gopher@golang.org", tenant, role, true)
func (c *Client) NewUser(name, password, email, tenantId, roleId string, enabled bool) (*User, error) {
	b := bytes.NewBufferString(fmt.Sprintf(`{"user": {"name": "%s", "password": "%s", "tenantId": "%s", "email": "%s", "enabled": %t}}`, name, password, tenantId, email, enabled))
	response, err := c.do("POST", c.authUrl+"/users", b)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result, _ := ioutil.ReadAll(response.Body)
	var data map[string]map[string]interface{}
	_ = json.Unmarshal(result, &data)
	user := User{
		Id:    data["user"]["id"].(string),
		Name:  data["user"]["name"].(string),
		Email: data["user"]["email"].(string),
	}
	response, err = c.do("PUT", c.authUrl+"/tenants/"+tenantId+"/users/"+user.Id+"/roles/OS-KSADM/"+roleId, nil)
	if err != nil {
		panic(err)
	}
	return &user, nil
}

// NewEc2 generate a new EC2 credentials pair for the given user in the given
// tenant.
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

// RemoveEc2 removes an EC2 credentials pair from a giver user. To remove an
// EC2 credential, you need to provide the user that owns it and the access key
// (the secret key is not necessary).
func (c *Client) RemoveEc2(userId, access string) error {
	return c.delete(c.authUrl + "/users/" + userId + "/credentials/OS-EC2/" + access)
}

// RemoveUser removes a user. You need also to provide a tenant and a role, so
// the role can be removed before the user is deleted.
func (c *Client) RemoveUser(userId, tenantId, roleId string) error {
	// FIXME(fsouza): deal with errors. Keystone keep returning malformed response.
	c.delete(c.authUrl + "/tenants/" + tenantId + "/users/" + userId + "/roles/OS-KSADM/" + roleId)
	return c.delete(c.authUrl + "/users/" + userId)
}

// RemoveTenant removes a tenant by its id.
func (c *Client) RemoveTenant(tenantId string) error {
	// FIXME(fsouza): deal with errors. Keystone keep returning malformed response.
	c.delete(c.authUrl + "/tenants/" + tenantId)
	return nil
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
