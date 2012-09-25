// Package testing provides a fake HTTP server, that can be user for tests.
package testing

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const ChanSize = 64

// TestHTTPServer holds information about the server.
type TestHTTPServer struct {
	// The URL of the server.
	URL string

	started  bool
	Request  chan *http.Request
	response chan *testResponse
	body     chan []byte
	timeout  time.Duration
}

type testResponse struct {
	Status  int
	Headers map[string]string
	Body    string
}

// NewTestHTTPServer returns an instance of TestHTTPServer.
//
// The given url is where the server will listen, and the responseTimeout
// parameter is how much time the server will wait before abort a request.
//
// Example of call:
//
//     NewTestHTTPServer("http://localhost:9898", 10e9)
//
// The call above will prepare a server that when started will listen for
// requests in the port 9898 (you should start it using Start method).
func NewTestHTTPServer(url string, responseTimeout time.Duration) *TestHTTPServer {
	return &TestHTTPServer{URL: url, timeout: responseTimeout}
}

// Start starts the server in the url provided to the NewTestHTTPServer call.
//
// It waits until the server is up.
func (s *TestHTTPServer) Start() {
	if s.started {
		return
	}
	s.started = true
	s.Request = make(chan *http.Request, ChanSize)
	s.response = make(chan *testResponse, ChanSize)
	s.body = make(chan []byte, ChanSize)
	url, _ := url.Parse(s.URL)
	go http.ListenAndServe(url.Host, s)
	s.PrepareResponse(202, nil, "Nothing.")
	for {
		resp, err := http.Get(s.URL)
		if err == nil && resp.StatusCode == 202 {
			break
		}
		time.Sleep(1e8)
	}
	s.WaitRequest(1e18)
}

// FlushRequests discards all requests. It is useful to call this method in the
// end of the test (teardown), so requests from a test does not affect the
// result of other test.
func (s *TestHTTPServer) FlushRequests() {
	for {
		select {
		case <-s.Request:
		default:
			return
		}
	}
}

// ServeHTTP is declared to make TestHTTPServer implement the http.Handler
// interface. It will handle all requests, storing information about them. You
// can use PrepareResponse method to prepare the next response, and WaitRequest
// to wait for a request (so you can test attributes of the request).
func (s *TestHTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	b, _ := ioutil.ReadAll(req.Body)
	s.body <- b
	s.Request <- req
	var resp *testResponse
	select {
	case resp = <-s.response:
	case <-time.After(s.timeout):
		log.Panicf("No response from server after %s.", s.timeout)
	}
	if resp.Status != 0 {
		w.WriteHeader(resp.Status)
	}
	w.Write([]byte(resp.Body))
}

// WaitRequest waits at most `timeout` for a request to be sent to the server.
//
// It returns the request pointer, the body of the request and a nil error in
// case of success, or an error if the operation times out.
//
// Example of call:
//
//     req, body, err := server.WaitRequest(10e9)
//
// The call above will wait at most 10 seconds for a request to come to the
// server.
func (s *TestHTTPServer) WaitRequest(timeout time.Duration) (*http.Request, []byte, error) {
	select {
	case req := <-s.Request:
		b := <-s.body
		return req, b, nil
	case <-time.After(timeout):
	}
	return nil, nil, errors.New("timed out")
}

// PrepareResponse should be called to prepare a response for a request.
// TestHTTPServer keeps an internal queue of responses, so sequent calls of
// PrepareResponse just add response to this queue.
//
// For example, the following calls will prepare three responses and consume
// them in the same sequence that they were prepared:
//
//     server.PrepareResponse(200, nil, "success")
//     server.PrepareResponse(200, map[string]string{"Content-Type": "application/json"}, `{"message":"success"}`)
//     server.PrepareResponse(204, nil, "")
//     http.Get(server.URL) // Gets the first response (200 - "succes").
//     http.Get(server.URL) // Gets the second response (200 - json {"message":"success"}).
//     http.Get(server.URL) // Gets the third response (204 - no content).
//     http.Get(server.URL) // Times out.
//
// Notice that a request in the server without a prepared response will timeout
// after a specific time (defined in the NewTestHTTPServer function).
func (s *TestHTTPServer) PrepareResponse(status int, headers map[string]string, body string) {
	s.response <- &testResponse{status, headers, body}
}
