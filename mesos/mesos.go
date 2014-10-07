package mesos

import (
	"encoding/json"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"log"
)

type MesosSlave struct {
	Slave      string
	HttpClient *httpclient.HttpClient
}

type SlaveState struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
}

func (slave *MesosSlave) State() SlaveState {
	response, err := slave.HttpClient.Get(fmt.Sprintf("http://%s/slave(1)/state.json", slave.Slave), nil)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	state := SlaveState{}
	err = decoder.Decode(&state)

	if err != nil {
		log.Fatal(err)
	}
	return state
}
