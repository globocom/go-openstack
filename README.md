Go OpenStack client
===================

[![Build Status](https://secure.travis-ci.org/timeredbull/openstack.png?branch=master)](http://travis-ci.org/timeredbull/openstack)

This is a go client for the OpenStack APIs.

Currently it works with Keystone 2.0 API and Nova API (in keystone and nova
subpackages).

By way of a quick-start, Keystone client:

```go
// use v2.0 auth with http://example.com:35357/v2.0")
client, err := NewClient("username", "pass", "admin", "http://example.com:35357/v2.0")
tenant, err := client.NewTenant("name", "desc", true)
client.RemoveTenant(tenant.Id)
```
