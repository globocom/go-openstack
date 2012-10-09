// Copyright 2012 go-openstack authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package keystone

import (
	ostesting "github.com/globocom/go-openstack/testing"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"testing"
)

var _ = Suite(&S{})

type S struct {
	response       string
	brokenResponse string
}

func Test(t *testing.T) { TestingT(t) }

var testServer = ostesting.NewTestHTTPServer("http://localhost:4444", 10e9)

func (s *S) SetUpSuite(c *C) {
	testServer.Start()
	body, err := ioutil.ReadFile("testdata/response.json")
	c.Assert(err, IsNil)
	s.response = string(body)
	brokenBody, err := ioutil.ReadFile("testdata/broken_response.json")
	c.Assert(err, IsNil)
	s.brokenResponse = string(brokenBody)
}

func (s *S) TearDownTest(c *C) {
	testServer.FlushRequests()
}
