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
	client, err := NewClient("username", "bad_pass", "tenantname", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	c.Assert(client, DeepEquals, &Client{Token: "secret"})
}
