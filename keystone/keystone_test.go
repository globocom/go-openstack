package keystone

import (
	. "launchpad.net/gocheck"
	"net/http"
)

func (s *S) TestAuthFailure(c *C) {
	testServer.PrepareResponse(401, nil, `{"error": {"message": "Invalid user / password", "code": 401, "title": "Not Authorized"}}`)
	client, err := NewClient("username", "bad_pass", "tenantname", "http://localhost:4444")
	c.Assert(client, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "Not Authorized")
}

func (s *S) TestAuth(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "tenantname", testServer.URL)
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	c.Assert(client.Token, Equals, "secret")
	c.Assert(client.authUrl, Equals, "http://localhost:4444")
	c.Assert(client.Catalogs, HasLen, 7)
	c.Assert(client.Catalogs[0].Name, Equals, "Compute Service")
	c.Assert(client.Catalogs[0].Type, Equals, "compute")
}

func (s *S) TestEndpoint(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "tenantname", testServer.URL)
	c.Assert(err, IsNil)
	c.Assert(client.Endpoint("compute", "admin"), Equals, "http://nova.mycloud.com:8774/v2/xpto")
	c.Assert(client.Endpoint("compute", "adminURL"), Equals, "http://nova.mycloud.com:8774/v2/xpto")
	c.Assert(client.Endpoint("sempute", "admin"), Equals, "")
}

func (s *S) TestNewTenant(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
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
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "admin", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	testServer.PrepareResponse(200, nil, `{"user": {"id": "userId", "enabled": "true", "name": "Stark", "email": "stark@stark.com"}}`)
	testServer.PrepareResponse(200, nil, "")
	user, err := client.NewUser("Stark", "mypass", "stark@stark.com", "mytenant", "member123", true)
	c.Assert(err, IsNil)
	c.Assert(user, NotNil)
	c.Assert(user, DeepEquals, &User{Id: "userId", Name: "Stark", Email: "stark@stark.com"})
}

func (s *S) TestNewEc2(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "admin", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	testServer.PrepareResponse(200, nil, `{"credential": {"access": "access", "secret": "secret"}}`)
	ec2, err := client.NewEc2("user", "tenant")
	c.Assert(err, IsNil)
	c.Assert(ec2, NotNil)
	c.Assert(ec2, DeepEquals, &Ec2{Access: "access", Secret: "secret"})
}

func (s *S) TestUserAddRole(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "admin", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	testServer.FlushRequests()
	testServer.PrepareResponse(200, nil, `"{role": {"id": "role-uuid-1234", "name": "rolename"}}`)
	err = client.AddRoleToUser("user-uuid4321", "tenant-uuid-567", "role-uuid-1234")
	c.Assert(err, IsNil)
	var request *http.Request
	request = <-testServer.Request
	expectedUrl := "/tenants/tenant-uuid-567/users/user-uuid4321/roles/OS-KSADM/role-uuid-1234"
	c.Assert(request.URL.Path, Equals, expectedUrl)
}

func (s *S) TestRemoveEc2(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "admin", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	testServer.PrepareResponse(200, nil, "")
	err = client.RemoveEc2("user", "access")
	c.Assert(err, IsNil)
}

func (s *S) TestRemoveEc2ReturnErrorIfItFailsToRemoveCredentials(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, _ := NewClient("username", "pass", "admin", "http://localhost:4444")
	testServer.PrepareResponse(500, nil, "Failed to remove credential.")
	err := client.RemoveEc2("stark123", "access-key")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^Failed to remove credential.$")
}

func (s *S) TestRemoveUser(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "admin", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	testServer.PrepareResponse(200, nil, "")
	testServer.PrepareResponse(200, nil, "")
	err = client.RemoveUser("user", "tenant", "member123")
	c.Assert(err, IsNil)
}

func (s *S) TestRemoveUserReturnErrorIfItFailsToRemoveUser(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, _ := NewClient("username", "pass", "admin", "http://localhost:4444")
	testServer.PrepareResponse(200, nil, "")
	testServer.PrepareResponse(500, nil, "Failed to remove user.")
	err := client.RemoveUser("start123", "tenant", "member123")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^Failed to remove user.$")
}

func (s *S) TestRemoveUserReturnErrorIfItFailsToRemoveTheRole(c *C) {
	// TODO(fsouza): re-enable this test when keystone start work like a real
	// HTTP server (see the FIXME note in the RemoveUser function).
	c.SucceedNow()
	testServer.PrepareResponse(200, nil, s.response)
	client, _ := NewClient("username", "pass", "admin", "http://localhost:4444")
	testServer.PrepareResponse(500, nil, "Failed to remove the role.")
	err := client.RemoveUser("start123", "tenant", "member123")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^Failed to remove the role.$")
}

func (s *S) TestRemoveTenant(c *C) {
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "admin", "http://localhost:4444")
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	body := `{"tenant": {"id": "xpto", "enabled": "true", "name": "name", "description": "desc"}}`
	testServer.PrepareResponse(200, nil, body)
	tenant, err := client.NewTenant("name", "desc", true)
	c.Assert(err, IsNil)
	c.Assert(tenant, NotNil)
	testServer.PrepareResponse(200, nil, "")
	err = client.RemoveTenant(tenant.Id)
	c.Assert(err, IsNil)
}

func (s *S) TestRemoveTenantReturnErrorIfItFailsToRemoveATenant(c *C) {
	// TODO(fsouza): re-enable this test when keystone start work like a real
	// HTTP server (see the FIXME note in the RemoveTenant function).
	c.SucceedNow()
	testServer.PrepareResponse(200, nil, s.response)
	client, err := NewClient("username", "pass", "admin", testServer.URL)
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)
	testServer.PrepareResponse(500, nil, "Failed to delete tenant.")
	err = client.RemoveTenant("uuid123")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^Failed to delete tenant.$")
}
