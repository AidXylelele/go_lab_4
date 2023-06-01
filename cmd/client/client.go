package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var target = flag.String("target", "http://localhost:8090", "request target")
var responseSize = flag.Int("size", 2023, "desired server response size")

func main() {
	flag.Parse()
	client := new(http.Client)
	client.Timeout = 10 * time.Second

	for range time.Tick(1 * time.Second) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/some-data", *target), nil)
		if err != nil {
			log.Printf("error creating request: %s", err)
			continue
		}
		req.Header.Set("Response-Size", strconv.Itoa(*responseSize))

		resp, err := client.Do(req)
		if err == nil {
			log.Printf("response %d", resp.StatusCode)
		} else {
			log.Printf("error %s", err)
		}
	}
}
