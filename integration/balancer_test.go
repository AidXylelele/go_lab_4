package integration

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	. "gopkg.in/check.v1"
)

const baseAddress = "http://balancer:8090"

type Client struct {
	client http.Client
}

func NewClient() *Client {
	return &Client{
		client: http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Client) Get(responseSize int) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/some-data", baseAddress), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Response-Size", strconv.Itoa(responseSize))

	return c.client.Do(req)
}

type IntegrationSuite struct {
	client *Client
}

func (s *IntegrationSuite) SetUpSuite(c *C) {
	s.client = NewClient()
}

func (s *IntegrationSuite) TestBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	// test server1
	server1, err := s.client.Get(3072)
	if err != nil {
		c.Error(err)
	}

	server1Header := server1.Header.Get("lb-from")
	c.Check(server1Header, Equals, "server1:8080")

	// test server2
	server2, err := s.client.Get(1024)
	if err != nil {
		c.Error(err)
	}

	server2Header := server2.Header.Get("lb-from")
	c.Check(server2Header, Equals, "server2:8080")

	// test server3
	server3, err := s.client.Get(2048)
	if err != nil {
		c.Error(err)
	}

	server3Header := server3.Header.Get("lb-from")
	c.Check(server3Header, Equals, "server3:8080")
	// again server 2
	two_server2, err := s.client.Get(1024)
	if err != nil {
		c.Error(err)
	}

	two_server2Header := two_server2.Header.Get("lb-from")
	c.Check(two_server2Header, Equals, "server2:8080")

	// test server3 again
	two_server3, err := s.client.Get(1024)
	if err != nil {
		c.Error(err)
	}

	two_server3Header := two_server3.Header.Get("lb-from")
	c.Check(two_server3Header, Equals, "server1:8080")
}

func (s *IntegrationSuite) BenchmarkBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	for i := 0; i < c.N; i++ {
		_, err := s.client.Get(10)
		if err != nil {
			c.Error(err)
		}
	}
}
