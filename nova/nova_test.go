// Copyright 2012 go-openstack authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nova

import (
	"github.com/globocom/go-openstack/keystone"
	ostesting "github.com/globocom/go-openstack/testing"
	. "launchpad.net/gocheck"
	"testing"
)

type S struct{}

var _ = Suite(&S{})

func Test(t *testing.T) {
	TestingT(t)
}

var testServer = ostesting.NewTestHTTPServer("http://localhost:5555", 1e9)

func (s *S) SetUpSuite(c *C) {
	testServer.Start()
}

func (s *S) TearDownTest(c *C) {
	testServer.FlushRequests()
}

func (s *S) TestDisassociateNetwork(c *C) {
	kclient := keystone.Client{
		Token: "123token",
		Catalogs: []keystone.ServiceCatalog{
			keystone.ServiceCatalog{
				Name: "Compute Service",
				Type: "compute",
				Endpoints: []map[string]string{
					map[string]string{
						"adminURL": "http://localhost:5555/v2/123tenant",
					},
				},
			},
		},
	}
	body := `{"networks": [{"bridge": "br1808", "vpn_public_port": 1000, "dhcp_start": "172.25.8.3", "bridge_interface": "eth1", "updated_at": "2012-05-12 02:16:48", "id": "ef0aa0c4-48d8-4d9e-903a-61486cd60805", "cidr_v6": null, "deleted_at": null, "gateway": "172.25.8.1", "label": "private_0", "project_id": "123tenant", "vpn_private_address": "172.25.8.2", "deleted": false, "vlan": 1808, "broadcast": "172.25.8.255", "netmask": "255.255.255.0", "injected": false, "cidr": "172.25.8.0/24", "vpn_public_address": "10.170.0.14", "multi_host": true, "dns1": null, "host": null, "gateway_v6": null, "netmask_v6": null, "created_at": "2012-05-12 02:13:17"}, {"bridge": "br1808", "vpn_public_port": 1000, "dhcp_start": "172.25.8.3", "bridge_interface": "eth1", "updated_at": "2012-05-12 02:16:48", "id": "ef0aa0c5-48d8-4d9e-903a-61486cd60805", "cidr_v6": null, "deleted_at": null, "gateway": "172.25.8.1", "label": "private_0", "project_id": "1234tenant", "vpn_private_address": "172.25.8.2", "deleted": false, "vlan": 1808, "broadcast": "172.25.8.255", "netmask": "255.255.255.0", "injected": false, "cidr": "172.25.8.0/24", "vpn_public_address": "10.170.0.14", "multi_host": true, "dns1": null, "host": null, "gateway_v6": null, "netmask_v6": null, "created_at": "2012-05-12 02:13:17"}]}`
	testServer.PrepareResponse(200, map[string]string{"Content-Type": "application/json"}, body) // List networks
	testServer.PrepareResponse(202, nil, "")                                                     // Disassociate network
	client := Client{KeystoneClient: &kclient}
	err := client.DisassociateNetwork("123tenant")
	c.Assert(err, IsNil)
	listreq, _, err := testServer.WaitRequest(1e9)
	if err != nil {
		c.Error("Did not send the request to retrieve the list of networks after 1 second.")
		c.FailNow()
	}
	disreq, b, err := testServer.WaitRequest(1e9)
	if err != nil {
		c.Error("Did not send the request to disassociate the network after 1 second.")
		c.FailNow()
	}
	c.Assert(err, IsNil)
	c.Assert(listreq.URL.Path, Equals, "/v2/123tenant/os-networks")
	c.Assert(listreq.Header.Get("X-Auth-Token"), Equals, "123token")
	c.Assert(disreq.URL.Path, Equals, "/v2/123tenant/os-networks/ef0aa0c4-48d8-4d9e-903a-61486cd60805/action")
	c.Assert(disreq.Header.Get("X-Auth-Token"), Equals, "123token")
	c.Assert(disreq.Header.Get("Content-Type"), Equals, "application/json")
	c.Assert(disreq.Header.Get("Accept"), Equals, "application/json")
	c.Assert(string(b), Equals, `{"disassociate":null}`)
}

func (s *S) TestDisassociateNetworkForTenantWithoutNetwork(c *C) {
	kclient := keystone.Client{
		Token: "123token",
		Catalogs: []keystone.ServiceCatalog{
			keystone.ServiceCatalog{
				Name: "Compute Service",
				Type: "compute",
				Endpoints: []map[string]string{
					map[string]string{
						"adminURL": "http://localhost:5555/v2/123tenant",
					},
				},
			},
		},
	}
	body := `{"networks": [{"bridge": "br1808", "vpn_public_port": 1000, "dhcp_start": "172.25.8.3", "bridge_interface": "eth1", "updated_at": "2012-05-12 02:16:48", "id": "ef0aa0c4-48d8-4d9e-903a-61486cd60805", "cidr_v6": null, "deleted_at": null, "gateway": "172.25.8.1", "label": "private_0", "project_id": "123tenant", "vpn_private_address": "172.25.8.2", "deleted": false, "vlan": 1808, "broadcast": "172.25.8.255", "netmask": "255.255.255.0", "injected": false, "cidr": "172.25.8.0/24", "vpn_public_address": "10.170.0.14", "multi_host": true, "dns1": null, "host": null, "gateway_v6": null, "netmask_v6": null, "created_at": "2012-05-12 02:13:17"}, {"bridge": "br1808", "vpn_public_port": 1000, "dhcp_start": "172.25.8.3", "bridge_interface": "eth1", "updated_at": "2012-05-12 02:16:48", "id": "ef0aa0c5-48d8-4d9e-903a-61486cd60805", "cidr_v6": null, "deleted_at": null, "gateway": "172.25.8.1", "label": "private_0", "project_id": "1234tenant", "vpn_private_address": "172.25.8.2", "deleted": false, "vlan": 1808, "broadcast": "172.25.8.255", "netmask": "255.255.255.0", "injected": false, "cidr": "172.25.8.0/24", "vpn_public_address": "10.170.0.14", "multi_host": true, "dns1": null, "host": null, "gateway_v6": null, "netmask_v6": null, "created_at": "2012-05-12 02:13:17"}]}`
	testServer.PrepareResponse(200, map[string]string{"Content-Type": "application/json"}, body) // List networks
	client := Client{KeystoneClient: &kclient}
	err := client.DisassociateNetwork("123tenantsojfdkw")
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNoNetwork)
}

func (s *S) TestDisassociateNetworkWithoutKeystoneClient(c *C) {
	cli := Client{}
	err := cli.DisassociateNetwork("anything")
	c.Assert(err, NotNil)
}
