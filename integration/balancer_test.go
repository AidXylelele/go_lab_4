package integration

import (
	"encoding/json"
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
const key = "vns-2023"
const responseSize1 = 1000
const responseSize2 = 2000
const responseSize3 = 3000

type RespBody struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var client = http.Client{
	Timeout: 3 * time.Second,
}

type IntegrationTestSuite struct{}

var _ = Suite(&IntegrationTestSuite{})

func TestBalancer(t *testing.T) {
	TestingT(t)
}

func sendRequest(baseAddress string, responseSize int, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/some-data", baseAddress), nil)
	if err != nil {
		log.Printf("error creating request: %s", err)
		return nil, err
	}
	req.Header.Set("Response-Size", strconv.Itoa(responseSize))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error: %s", err)
		return nil, err
	}

	log.Printf("response %d", resp.StatusCode)
	return resp, nil
}

func (s *IntegrationTestSuite) TestGetRequest(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}
	server1, _ := sendRequest(baseAddress, responseSize3, &client)
	c.Check(server1.Header.Get("lb-from"), Equals, "server1:8080")

	server2, _ := sendRequest(baseAddress, responseSize2, &client)
	c.Check(server2.Header.Get("lb-from"), Equals, "server2:8080")

	server3, _ := sendRequest(baseAddress, responseSize1, &client)
	c.Check(server3.Header.Get("lb-from"), Equals, "server3:8080")

	server3_again, _ := sendRequest(baseAddress, responseSize3, &client)
	c.Check(server3_again.Header.Get("lb-from"), Equals, "server3:8080")

	server2_again, _ := sendRequest(baseAddress, responseSize2, &client)
	c.Check(server2_again.Header.Get("lb-from"), Equals, "server2:8080")

	server1_again, _ := sendRequest(baseAddress, responseSize1, &client)
	c.Check(server1_again.Header.Get("lb-from"), Equals, "server1:8080")

	db, err := client.Get(fmt.Sprintf("%s/api/v1/some-data?key=%s", baseAddress, key))
	if err != nil {
		c.Error(err)
	}
	var body RespBody
	err = json.NewDecoder(db.Body).Decode(&body)
	if err != nil {
		c.Error(err)
	}
	c.Check(body.Key, Equals, "vns-2023")
	if body.Value == "" {
		c.Error(err)
	}
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
