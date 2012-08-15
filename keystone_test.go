package keystone

import (
	. "launchpad.net/gocheck"
)

func (s *S) TestAuthFailure(c *C) {
	testServer.PrepareResponse(401, nil, `{"error": {"message": "Invalid user / password", "code": 401, "title": "Not Authorized"}}`)
	client, err := NewClient("username", "bad_pass", "tenantname", "http://localhost:4444")
	c.Assert(client, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "Not Authorized")
}

func (s *S) TestAuth(c *C) {
	testServer.PrepareResponse(200, nil, `{"access": {"token": {"id": "secret"}}}`)
	client, err := NewClient("username", "pass", "tenantname", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	c.Assert(client, DeepEquals, &Client{Token: "secret", authUrl: "http://localhost:4444"})
}

func (s *S) TestNewTenant(c *C) {
	testServer.PrepareResponse(200, nil, `{"access": {"token": {"id": "secret"}}}`)
	client, err := NewClient("username", "pass", "admin", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	testServer.PrepareResponse(200, nil, `{"tenant": {"id": "xpto", "enabled": "true", "name": "name", "description": "desc"}}`)
	tenant, err := client.NewTenant("name", "desc", true)
	c.Assert(err, IsNil)
	c.Assert(tenant, NotNil)
	c.Assert(tenant, DeepEquals, &Tenant{Id: "xpto", Name: "name", Description: "desc"})
}

func (s *S) TestNewUser(c *C) {
	testServer.PrepareResponse(200, nil, `{"access": {"token": {"id": "secret"}}}`)
	client, err := NewClient("username", "pass", "admin", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	testServer.PrepareResponse(200, nil, `{"tenant": {"id": "xpto", "enabled": "true", "name": "name", "description": "desc"}}`)
	tenant, err := client.NewTenant("name", "desc", true)
	c.Assert(err, IsNil)
	c.Assert(tenant, NotNil)
	testServer.PrepareResponse(200, nil, `{"user": {"id": "userId", "enabled": "true", "name": "Stark", "email": "stark@stark.com"}}`)
	user, err := client.NewUser("Stark", "mypass", "stark@stark.com", tenant.Id, true)
	c.Assert(err, IsNil)
	c.Assert(user, NotNil)
	c.Assert(user, DeepEquals, &User{Id: "userId", Name: "Stark", Email: "stark@stark.com"})
}
