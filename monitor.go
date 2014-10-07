package main

import (
	"flag"
	"github.com/armon/consul-api"
	"github.com/chrono/marathon-consul-discovery/consul-marathon"
	"github.com/chrono/marathon-consul-discovery/marathon"
	"github.com/chrono/marathon-consul-discovery/mesos"
	"github.com/ddliu/go-httpclient"
)

var bind = flag.String("bind", "0.0.0.0", "address to listen on and register with marathon -- 0.0.0.0 auto discovers via mesos slave")
var port = flag.Int("port", 8080, "http port to listen on")
var marathon_endpoint = flag.String("marathon", "zookeeper:8080", "marathon to register with")
var mesos_slave = flag.String("mesos_slave", "localhost:5051", "mesos slave to handle tasks for")

func main() {
	flag.Parse()

	http_client := httpclient.NewHttpClient(nil)

	var my_marathon = marathon.Marathon{
		Master:     *marathon_endpoint,
		HttpClient: http_client,
	}

	var slave = mesos.MesosSlave{
		Slave:      *mesos_slave,
		HttpClient: http_client,
	}

	var consul, _ = consulapi.NewClient(consulapi.DefaultConfig())

	var listen_address = *bind
	if listen_address == "0.0.0.0" {
		listen_address = slave.State().Hostname
	}

	go consulmarathon.PollMarathonTasks(my_marathon, consul.Agent(), slave.State().Hostname)
	handler := marathon.MarathonEventHandler{
		Marathon: my_marathon,
		Address:  listen_address,
		Port:     *port,
		Events:   make(chan interface{}),
	}
	go consulmarathon.ProcessMarathonConsulEvents(handler.Events, slave.State().Id, consul.Agent())
	handler.SubscribeEvents()
}
