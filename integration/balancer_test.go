package integration

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

const baseAddress = "http://balancer:8090"

var responseSize = flag.Int("size", 2023, "desired server response size")

var client = http.Client{
	Timeout: 3 * time.Second,
}

type IntegrationTestSuite struct{}

var _ = Suite(&IntegrationTestSuite{})

func TestBalancer(t *testing.T) {
	TestingT(t)
}

func (s *IntegrationTestSuite) TestGetRequest(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/some-data", baseAddress), nil)
	if err != nil {
		log.Printf("error creating request: %s", err)
	}
	req.Header.Set("Response-Size", strconv.Itoa(*responseSize))

	resp, err := client.Do(req)
	if err == nil {
		log.Printf("response %d", resp.StatusCode)
	} else {
		log.Printf("error %s", err)
	}

	c.Check(resp.Header.Get("lb-from"), Equals, "server1:8080")

}

func (s *IntegrationTestSuite) BenchmarkBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	for i := 0; i < c.N; i++ {
		_, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {
			c.Error(err)
		}
	}
}
