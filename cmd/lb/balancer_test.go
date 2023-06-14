package main

import (
	"testing"
	"time"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type BalancerSuite struct{}

var _ = check.Suite(&BalancerSuite{})

func (s *BalancerSuite) TestBalancer(c *check.C) {
	healthChecker := &HealthChecker{}
	healthChecker.healthyServers = serversPool

	balancer := &Balancer{}
	balancer.healthChecker = healthChecker

	//seting serverLoads
	balancer.updateLowestLoadIndex(map[string]int64{
		"server1:8080": 100,
		"server2:8080": 200,
		"server3:8080": 150,
	})

	server1 := balancer.getServerWithLowestLoad()

	//seting serverLoads
	balancer.updateLowestLoadIndex(map[string]int64{
		"server1:8080": 300,
		"server2:8080": 200,
		"server3:8080": 250,
	})

	server2 := balancer.getServerWithLowestLoad()

	//seting serverLoads
	balancer.updateLowestLoadIndex(map[string]int64{
		"server1:8080": 200,
		"server2:8080": 150,
		"server3:8080": 100,
	})

	server3 := balancer.getServerWithLowestLoad()

	c.Assert(server1, check.Equals, serversPool[0])
	c.Assert(server2, check.Equals, serversPool[1])
	c.Assert(server3, check.Equals, serversPool[2])
}

func (s *BalancerSuite) TestHealthChecker(c *check.C) {
	healthChecker := &HealthChecker{}
	healthChecker.health = func(s string) bool {
		if s == "1" {
			return false
		} else {
			return true
		}
	}

	healthChecker.serversPool = []string{"1", "2", "3"}
	healthChecker.healthyServers = []string{"4", "5", "6"}
	healthChecker.checkInterval = 1 * time.Second

	healthChecker.StartHealthCheck()

	time.Sleep(2 * time.Second)

	c.Assert(healthChecker.healthyServers[0], check.Equals, "2")
	c.Assert(healthChecker.healthyServers[1], check.Equals, "3")
	c.Assert(len(healthChecker.healthyServers), check.Equals, 2)
}
