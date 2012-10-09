// Copyright 2012 go-openstack authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"fmt"
	"github.com/globocom/go-openstack/keystone"
	"os"
)

var authUrl = getEnv("KEYSTONE_AUTH_URL")
var userName = getEnv("KEYSTONE_USER")
var tenantName = getEnv("KEYSTONE_TENANT")
var password = getEnv("KEYSTONE_PASSWORD")
var memberRole = getEnv("KEYSTONE_MEMBER_ROLE")
var staticRole = getEnv("KEYSTONE_STATIC_ROLE")

func getEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		fmt.Printf("You must define the environment variable %s.\n", name)
		os.Exit(3)
	}
	return val
}

func tearDown(client *keystone.Client, userId, access, tenantId string) {
	fmt.Println("tearing down....")
	client.RemoveEc2(userId, access)
	client.RemoveRoleFromUser(tenantId, userId, staticRole)
	client.RemoveUser(userId, tenantId, memberRole)
	client.RemoveTenant(tenantId)
}

func main() {
	client, err := keystone.NewClient(userName, password, tenantName, authUrl)
	if err != nil {
		panic(err)
	}

	tenant, err := client.NewTenant("smoketests", "smoking", true)
	if err != nil {
		panic("Failed to create a tenant: " + err.Error())
	}
	fmt.Println("Tenant smoketests created.")

	user, err := client.NewUser("smoketests", "smoketests", "smoketests@tsuru.org", tenant.Id, memberRole, true)
	if err != nil {
		tearDown(client, user.Id, "", tenant.Id)
		panic("Failed to create a user: " + err.Error())
	}
	fmt.Println("User smoketests created.")

	ec2, err := client.NewEc2(user.Id, tenant.Id)
	if err != nil {
		tearDown(client, user.Id, ec2.Access, tenant.Id)
		panic("Failed to create ec2 creds: " + err.Error())
	}
	fmt.Println("Credentials for user smoketests generated.")

	err = client.AddRoleToUser(tenant.Id, user.Id, staticRole)
	if err != nil {
		tearDown(client, user.Id, ec2.Access, tenant.Id)
		panic("Failed to add role: " + err.Error())
	}
	fmt.Println("Added role to user.\n")

	err = client.RemoveRoleFromUser(tenant.Id, user.Id, staticRole)
	if err != nil {
		tearDown(client, user.Id, ec2.Access, tenant.Id)
		panic("Failed to remove role: " + err.Error())
	}
	fmt.Println("Removed role from user.")

	err = client.RemoveEc2(user.Id, ec2.Access)
	if err != nil {
		tearDown(client, user.Id, ec2.Access, tenant.Id)
		panic("Failed to remove ec2 creds: " + err.Error())
	}
	fmt.Println("Credentials for user smoketests removed.")

	err = client.RemoveUser(user.Id, tenant.Id, memberRole)
	if err != nil {
		tearDown(client, user.Id, ec2.Access, tenant.Id)
		panic("Failed to remove user: " + err.Error())
	}
	fmt.Println("User smoketests removed.")

	err = client.RemoveTenant(tenant.Id)
	if err != nil {
		tearDown(client, user.Id, ec2.Access, tenant.Id)
		panic("Failed to remove tenant: " + err.Error())
	}
	fmt.Println("Tenant smoketests removed.")
}
