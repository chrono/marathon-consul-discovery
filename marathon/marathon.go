package marathon

import (
	"encoding/json"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Marathon struct {
	Master     string
	HttpClient *httpclient.HttpClient
}

type MarathonInfo struct {
	FrameworkId string `json:"frameworkId"`
}

type MarathonTask struct {
	AppId     string `json:"appId"`
	Id        string `json:"id"`
	Host      string `json:"host"`
	Ports     []int  `json:"ports"`
	StartedAt string `json:"startedAt"`
	StagedAt  string `json:"stagedAt"`
	Version   string `json:"version"`
}

type MarathonTaskList struct {
	Tasks []MarathonTask `json:"tasks"`
}

func (marathon *Marathon) Info() MarathonInfo {
	response, err := marathon.HttpClient.Get(fmt.Sprintf("http://%s/v2/info", marathon.Master), nil)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	info := MarathonInfo{}

	body, _ := response.ReadAll()
	err = json.Unmarshal(body, &info)

	if err != nil {
		log.Fatal(err)
	}

	return info
}

func (marathon *Marathon) TaskList() MarathonTaskList {
	response, err := marathon.HttpClient.Get(fmt.Sprintf("http://%s/v2/tasks", marathon.Master), nil)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	task_list := MarathonTaskList{}

	body, _ := response.ReadAll()
	err = json.Unmarshal(body, &task_list)

	if err != nil {
		log.Fatal(err)
	}

	return task_list
}

func paramsToString(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	return values.Encode()
}

func (marathon *Marathon) updateSubscription(method string, callbackUri string) {
	subscribeUrl := url.URL{
		Scheme: "http",
		Host:   marathon.Master,
		Path:   "/v2/eventSubscriptions",
		RawQuery: paramsToString(map[string]string{
			"callbackUrl": callbackUri,
		}),
	}

	response, err := marathon.HttpClient.Do("POST", subscribeUrl.String(), map[string]string{"Accept": "application/json"}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Fatal(response.ToString())
	}
}

func (marathon *Marathon) Subscribe(callbackUri string) {
	marathon.updateSubscription("POST", callbackUri)
}

func (marathon *Marathon) Unsubscribe(callbackUri string) {
	marathon.updateSubscription("DELETE", callbackUri)
}

type StatusUpdateEvent struct {
	AppId      string `json:"appId"`
	Host       string `json:"host"`
	Ports      []int  `json:"ports"`
	SlaveId    string `json:"slaveId"`
	TaskId     string `json:"taskId"`
	TaskStatus string `json:"taskStatus"`
	Version    string `json:"version"`
}

type HealthStatusChangedEvent struct {
	Alive   bool   `json:"alive"`
	AppId   string `json:"appId"`
	Host    string `json:"host"` // host:port
	TaskId  string `json:"taskId"`
	Version string `json:"version"`
}

type MarathonEvent struct {
	EventType string `json:"eventType"`
}

type MarathonEventHandler struct {
	Marathon Marathon
	Address  string
	Port     int
	Events   chan interface{}
}

func (handler *MarathonEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sniff MarathonEvent
	var event interface{}
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &sniff)
	switch sniff.EventType {
	case "status_update_event":
		var e StatusUpdateEvent
		json.Unmarshal(body, &e)
		event = e
	case "health_status_changed_event": // not actually using this yet
		var e HealthStatusChangedEvent
		json.Unmarshal(body, &e)
		event = e
	}
	if event != nil {
		handler.Events <- event
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (handler *MarathonEventHandler) SubscribeEvents() {
	callbackUrl := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", handler.Address, handler.Port),
	}
	handler.Marathon.Subscribe(callbackUrl.String())
	// defer my_marathon.Unsubscribe(callbackUrl.String())
	http.Handle("/", handler)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", handler.Address, handler.Port), nil)
	if err != nil {
		log.Fatalln(err)
	}
}
