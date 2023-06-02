package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/AidXylelele/go_lab_4/httptools"
	"github.com/AidXylelele/go_lab_4/signal"
)

var port = flag.Int("port", 8080, "server port")

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

type HealthHandler struct{}

func (h HealthHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("content-type", "text/plain")
	if failConfig := os.Getenv(confHealthFailure); failConfig == "true" {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte("FAILURE"))
	} else {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("OK"))
	}
}

type SomeDataHandler struct{}

func (h SomeDataHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key != "" && r.Header.Get("Response-Size") != "" {
		if err := processDBRequest(rw, key); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		processDefaultRequest(rw, r)
	}
}

func processDBRequest(rw http.ResponseWriter, key string) error {
	client := http.DefaultClient
	resp, err := client.Get(fmt.Sprintf("%s/%s", dbUrl, key))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	statusOk := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOk {
		rw.WriteHeader(resp.StatusCode)
		return nil
	}

	respDelayString := os.Getenv(confResponseDelaySec)
	if delaySec, parseErr := strconv.Atoi(respDelayString); parseErr == nil && delaySec > 0 && delaySec < 300 {
		time.Sleep(time.Duration(delaySec) * time.Second)
	}

	var body RespBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	rw.Header().Set("content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(body); err != nil {
		return err
	}

	return nil
}

func processDefaultRequest(rw http.ResponseWriter, r *http.Request) {
	respDelayString := os.Getenv(confResponseDelaySec)
	if delaySec, parseErr := strconv.Atoi(respDelayString); parseErr == nil && delaySec > 0 && delaySec < 300 {
		time.Sleep(time.Duration(delaySec) * time.Second)
	}

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
}

type ReportHandler struct {
	report Report
}

func (h ReportHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(rw).Encode(h.report)
}

func main() {
	flag.Parse()

	h := http.NewServeMux()
	client := http.DefaultClient

	h.Handle("/health", HealthHandler{})
	report := make(Report)
	h.Handle("/api/v1/some-data", SomeDataHandler{})
	h.Handle("/report", ReportHandler{report: report})

	server := httptools.CreateServer(*port, h)
	server.Start()

	buff := new(bytes.Buffer)
	body := ReqBody{Value: time.Now().Format(time.RFC3339)}
	json.NewEncoder(buff).Encode(body)

	res, err := client.Post(fmt.Sprintf("%s/vns-2023", dbUrl), "application/json", buff)
	if err != nil {
		log.Printf("Error while sending test request to db: %s", err.Error())
	}
	defer res.Body.Close()

	signal.WaitForTerminationSignal()
}
