Go keystone client
==================

This is a go client for the OpenStack Keystone 2.0 API.

By way of a quick-start:

```go
# use v2.0 auth with http://example.com:35357/v2.0")
client, err := NewClient("username", "pass", "admin", "http://example.com:35357/v2.0")
tenant, err := client.NewTenant("name", "desc", true)
client.RemoveTenant(tenant.Id)
```
