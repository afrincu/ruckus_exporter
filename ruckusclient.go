package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

var (
	url = "https://YourSmartzoneIPAddress:8443/wsg/api/public/v6_1"
)

type RuckusClient http.Client

type Login struct {
	Name string `json:"username"`
	Pass string `json:"password"`
}

var login Login = Login{
	Name: "user",  // enter the username for the SmartZone controller
	Pass: "password", // enter the password for the SmartZone controller
}

type AP struct {
	Name        string `json:"deviceName"`
	Status      string `json:"status"`
	Clients5G   int    `json:"numClients5G"`
	Clients24G  int    `json:"numClients24G"`
	Noise5G     int    `json:"noise5G"`
	Noise24G    int    `json:"noise24G"`
	Airtime5G   int    `json:"airtime5G"`
	Airtime24G  int    `json:"airtime24G"`
	Latency5G   int    `json:"latency5G"`
	Latency24G  int    `json:"latency24G"`
	Retry5G     int    `json:"retry5G"`
	Retry24G    int    `json:"retry24G"`
	Capacity5G  int    `json:"capacity5G"`
	Capacity24G int    `json:"capacity24G"`
	Tx          int    `json:"tx"`
	Rx          int    `json:"rx"`
}
type MessageAP struct {
	Count int  `json:"totalCount"`
	Aps   []AP `json:"list"`
}
type WLAN struct {
	Name     string `json:"name"`
	ZoneName string `json:"zoneName"`
	Clients  int    `json:"clients"`
}
type MessageWLAN struct {
	Count int    `json:"totalCount"`
	WLANs []WLAN `json:"list"`
}

func NewRuckusClient() RuckusClient {
	cookieJar, _ := cookiejar.New(nil) // TODO
	return RuckusClient(http.Client{
		Jar:     cookieJar,
		Timeout: time.Second * 10,
	})
}

func (rc *RuckusClient) Login() error {
	account, err := json.Marshal(login)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url+"/session", bytes.NewReader(account))
	if err != nil {
		return err
	}

	resp, err := (*http.Client)(rc).Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login: unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (rc *RuckusClient) ListAP() (m MessageAP, err error) {
	limit := `{"attributes":["*"], "limit":100}`
	req, err := http.NewRequest(http.MethodPost, url+"/query/ap", strings.NewReader(limit))
	if err != nil {
		return
	}

	resp, err := (*http.Client)(rc).Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &m)
	if err != nil {
		return
	}

	return
}

func (rc *RuckusClient) WLAN() (m MessageWLAN, err error) {
	limit := `{"attributes":["*"], "limit":100}`
	req, err := http.NewRequest(http.MethodPost, url+"/query/wlan", strings.NewReader(limit))
	if err != nil {
		return
	}

	resp, err := (*http.Client)(rc).Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &m)
	if err != nil {
		return
	}

	return
}
