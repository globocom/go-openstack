package keystone

import (
	os_testing "github.com/timeredbull/openstack/testing"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"testing"
)

var _ = Suite(&S{})

type S struct {
	response string
}

func Test(t *testing.T) { TestingT(t) }

var testServer = os_testing.NewTestHTTPServer("http://localhost:4444")

func (s *S) SetUpSuite(c *C) {
	testServer.Start()
	body, err := ioutil.ReadFile("testdata/response.json")
	c.Assert(err, IsNil)
	s.response = string(body)
}

func (s *S) TearDownTest(c *C) {
	testServer.FlushRequests()
}
