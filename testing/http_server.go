package testing

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

type TestHTTPServer struct {
	URL      string
	started  bool
	request  chan *http.Request
	response chan *testResponse
}

type testResponse struct {
	Status  int
	Headers map[string]string
	Body    string
}

func NewTestHTTPServer(url string) *TestHTTPServer {
	return &TestHTTPServer{URL: url}
}

func (s *TestHTTPServer) Start() {
	if s.started {
		return
	}
	s.started = true
	s.request = make(chan *http.Request, 64)
	s.response = make(chan *testResponse, 64)
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

func (s *TestHTTPServer) FlushRequests() {
	for {
		select {
		case <-s.request:
		default:
			return
		}
	}
}

func (s *TestHTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.request <- req
	var resp *testResponse
	resp = <-s.response
	if resp.Status != 0 {
		w.WriteHeader(resp.Status)
	}
	w.Write([]byte(resp.Body))
}

func (s *TestHTTPServer) WaitRequest(timeout time.Duration) (*http.Request, error) {
	select {
	case req := <-s.request:
		req.ParseForm()
		return req, nil
	case <-time.After(timeout):
	}
	return nil, errors.New("timed out")
}

func (s *TestHTTPServer) PrepareResponse(status int, headers map[string]string, body string) {
	s.response <- &testResponse{status, headers, body}
}
