Go OpenStack client
===================

[![Build Status](https://secure.travis-ci.org/globocom/go-openstack.png?branch=master)](http://travis-ci.org/globocom/go-openstack)

This is a go client for the OpenStack APIs.

Currently it works with Keystone 2.0 API and Nova API (in keystone and nova
subpackages).

By way of a quick-start:

```go
// use v2.0 auth with http://example.com:35357/v2.0")
keystoneClient, err := keystone.NewClient("username", "pass", "admin", "http://example.com:35357/v2.0")
tenant, err := keystoneClient.NewTenant("name", "desc", true)
novaClient := nova.Client{KeystoneClient: keystoneClient}
novaClient.DisassociateNetwork(tenant.Id)
keystoneClient.RemoveTenant(tenant.Id)
```
