package keystone

import (
	. "launchpad.net/gocheck"
)

func (s *S) TestAuth(c *C) {
	testServer.PrepareResponse(401, nil, `{"error": {"message": "Invalid user / password", "code": 401, "title": "Not Authorized"}}`)
	_, err := NewClient("username", "bad_pass", "tenantname", "http://localhost:4444")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "Not Authorized")
}
