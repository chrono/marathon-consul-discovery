package consulmarathon

import (
	"fmt"
	"github.com/armon/consul-api"
	"github.com/chrono/marathon-consul-discovery/marathon"
	"log"
	"regexp"
	"strings"
	"time"
)

var consul_marathon_id_prefix = "marathon:"
var consul_marathon_name_prefix = ""
var dns_invalid_hostname, _ = regexp.Compile(`[^a-zA-Z0-9\-]`)
var dns_invalid_hostname_prefix, _ = regexp.Compile(`^[^a-zA-Z0-9\-]`)

func ConsulMarathonTaskFromEvent(event marathon.StatusUpdateEvent) ConsulMarathonTask {
	if len(event.Ports) == 0 {
		log.Fatalln("cannot create a consul task without a port", event)
	}
	return ConsulMarathonTask{
		marathon.MarathonTask{
			AppId:   event.AppId,
			Id:      event.TaskId,
			Host:    event.Host,
			Ports:   event.Ports,
			Version: event.Version,
		},
	}
}

type ConsulMarathonTask struct {
	marathon.MarathonTask
}

func (task *ConsulMarathonTask) ConsulId() string {
	return fmt.Sprintf("%s%s", consul_marathon_id_prefix, task.Id)
}

func (task *ConsulMarathonTask) ConsulCheckId() string {
	return fmt.Sprintf("service:%s", task.ConsulId())
}

func (task *ConsulMarathonTask) ConsulName() string {
	return fmt.Sprintf("%s%s", consul_marathon_name_prefix, dns_invalid_hostname.ReplaceAllString(dns_invalid_hostname_prefix.ReplaceAllString(task.AppId, ""), "-"))
}

func (task *ConsulMarathonTask) Registration() *consulapi.AgentServiceRegistration {
	return &consulapi.AgentServiceRegistration{
		ID:   task.ConsulId(),
		Name: task.ConsulName(),
		Port: task.Ports[0],
		Check: &consulapi.AgentServiceCheck{
			TTL: "15s",
		},
	}
}

func (task *ConsulMarathonTask) ensureRegistered(agent *consulapi.Agent, comment string) error {
	consul_services, _ := agent.Services()
	registration := task.Registration()
	if consul_services[registration.ID] == nil {
		err := agent.ServiceRegister(registration)
		if err != nil {
			return err
		}
	}

	return agent.PassTTL(task.ConsulCheckId(), comment)
}

func (task *ConsulMarathonTask) ensureAbsent(agent *consulapi.Agent) error {
	consul_services, _ := agent.Services()
	if consul_services[task.ConsulId()] != nil {
		return agent.ServiceDeregister(task.ConsulId())
	}
	return nil
}

func ProcessMarathonConsulEvents(consul_chan <-chan interface{}, my_slave_id string, agent *consulapi.Agent) {
	for {
		event := <-consul_chan
		switch event := event.(type) {
		case marathon.StatusUpdateEvent:
			if event.SlaveId == my_slave_id && len(event.Ports) > 0 {
				task := ConsulMarathonTaskFromEvent(event)
				switch event.TaskStatus {
				case "TASK_RUNNING":
					task.ensureRegistered(agent, "consulmarathon received StatusUpdateEvent with TaskStatus: TASK_RUNNING")
				case "TASK_KILLED", "TASK_FINISHED", "TASK_FAILED", "TASK_LOST":
					task.ensureAbsent(agent)
				}
			}
		}
	}
}

func PollMarathonTasks(my_marathon marathon.Marathon, agent *consulapi.Agent, hostname string) map[string]*marathon.MarathonTask {
	ticker := time.NewTicker(10 * time.Second)
	for {
		task_list := my_marathon.TaskList().Tasks
		my_task_map := make(map[string]*marathon.MarathonTask, 100)
		for _, task := range task_list {
			if hostname == task.Host && len(task.Ports) > 0 {
				task_copy := task
				my_task_map[task.Id] = &task_copy
			}
		}
		update_consul(agent, my_task_map)
		<-ticker.C
	}
}

func update_consul(agent *consulapi.Agent, marathon_tasks map[string]*marathon.MarathonTask) {
	consul_services, _ := agent.Services()

	// ensure tasks present in marathon are in consul
	for _, task := range marathon_tasks {
		if task.StartedAt != "" {
			task := ConsulMarathonTask{*task}
			task.ensureRegistered(agent, fmt.Sprintf("consulmarathon polled task from marathon api at %s", time.Now()))
		}
	}
	//
	// ensure tasks absent in marathon are absent in consul
	for _, service := range consul_services {
		if strings.HasPrefix(service.ID, consul_marathon_id_prefix) {
			if marathon_task_id := strings.TrimPrefix(service.ID, consul_marathon_id_prefix); marathon_tasks[marathon_task_id] == nil {
				task := ConsulMarathonTask{marathon.MarathonTask{Id: marathon_task_id}}
				task.ensureAbsent(agent)
			}
		}
	}
}
