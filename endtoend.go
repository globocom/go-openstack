// +build ignore

package main

import (
	"fmt"
	"github.com/timeredbull/keystone"
	"os"
)

func getEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		fmt.Printf("You must define the environment variable %s.\n", name)
		os.Exit(3)
	}
	return val
}

func main() {
	authUrl := getEnv("KEYSTONE_AUTH_URL")
	userName := getEnv("KEYSTONE_USER")
	tenantName := getEnv("KEYSTONE_TENANT")
	password := getEnv("KEYSTONE_PASSWORD")
	memberRole := getEnv("KEYSTONE_MEMBER_ROLE")
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
		panic("Failed to create a user: " + err.Error())
	}
	fmt.Println("User smoketests created.")
	ec2, err := client.NewEc2(user.Id, tenant.Id)
	if err != nil {
		panic("Failed to create ec2 creds: " + err.Error())
	}
	fmt.Println("Credentials for user smoketests generated.\n")
	err = client.RemoveEc2(user.Id, ec2.Access)
	if err != nil {
		panic("Failed to remove ec2 creds: " + err.Error())
	}
	fmt.Println("Credentials for user smoketests removed.")
	err = client.RemoveUser(user.Id, tenant.Id, memberRole)
	if err != nil {
		panic("Failed to remove user: " + err.Error())
	}
	fmt.Println("User smoketests removed.")
	err = client.RemoveTenant(tenant.Id)
	if err != nil {
		panic("Failed to remove tenant: " + err.Error())
	}
	fmt.Println("Tenant smoketests removed.")
}
