package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/AidXylelele/go_lab_4/httptools"
	"github.com/AidXylelele/go_lab_4/signal"
)

const (
	confResponseDelaySec = "CONF_RESPONSE_DELAY_SEC"
	confHealthFailure    = "CONF_HEALTH_FAILURE"
	dbUrl                = "http://db:8083/db"
)

type ReqBody struct {
	Value string `json:"value"`
}

type RespBody struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	client := http.DefaultClient

	http.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/plain")
		if failConfig := os.Getenv(confHealthFailure); failConfig == "true" {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("FAILURE"))
			return
		}
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("OK"))
	})

	report := make(Report)

	http.HandleFunc("/api/v1/some-data", func(rw http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key != "" {
			resp, err := client.Get(fmt.Sprintf("%s/%s", dbUrl, key))
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			if statusOk := resp.StatusCode >= 200 && resp.StatusCode < 300; !statusOk {
				rw.WriteHeader(resp.StatusCode)
				return
			}

			handleResponseDelay()

			report.Process(r)

			var body RespBody
			json.NewDecoder(resp.Body).Decode(&body)

			rw.Header().Set("content-type", "application/json")
			rw.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(rw).Encode(body)

			defer resp.Body.Close()
			return
		}

		handleResponseDelay()

		report.Process(r)

		rw.Header().Set("content-type", "application/json")
		rw.WriteHeader(http.StatusOK)

		responseSize := 1024 // Default response size
		if sizeHeader := r.Header.Get("Response-Size"); sizeHeader != "" {
			if size, err := strconv.Atoi(sizeHeader); err == nil && size > 0 {
				responseSize = size
			}
		}

		responseData := make([]string, responseSize)
		for i := 0; i < responseSize; i++ {
			responseData[i] = strconv.Itoa(responseSize)
		}

		_ = json.NewEncoder(rw).Encode(responseData)
	})

	h := http.NewServeMux()
	h.Handle("/report", report)

	server := httptools.CreateServer(8080, h)
	server.Start()

	buff := new(bytes.Buffer)
	body := ReqBody{Value: time.Now().Format(time.RFC3339)}
	json.NewEncoder(buff).Encode(body)

	res, err := client.Post(fmt.Sprintf("%s/test", dbUrl), "application/json", buff)
	if err != nil {
		log.Panic(err)
	}
	defer res.Body.Close()

	signal.WaitForTerminationSignal()
}

func handleResponseDelay() {
	respDelayString := os.Getenv(confResponseDelaySec)
	if delaySec, parseErr := strconv.Atoi(respDelayString); parseErr == nil && delaySec > 0 && delaySec < 300 {
		time.Sleep(time.Duration(delaySec) * time.Second)
	}
}
