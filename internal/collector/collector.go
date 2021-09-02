package collector

import (
	"fmt"
	"github.com/egeneralov/ejabberd_api_exporter/internal/api"
	"github.com/egeneralov/ejabberd_api_exporter/internal/generic/str"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	client *api.Api

	knownUsers []string

	initial bool

	descRegisteredUser *prometheus.Desc
	descConnectedUser  *prometheus.Desc
	descUserResources  *prometheus.Desc
	descUptime         *prometheus.Desc
	descProcesses      *prometheus.Desc
}

func New(client *api.Api, namespace string) *Collector {
	return &Collector{
		client:  client,
		initial: true,
		descRegisteredUser: prometheus.NewDesc(
			namespace+"_registered_user",
			"registered users list",
			[]string{"username"},
			nil,
		),
		descConnectedUser: prometheus.NewDesc(
			namespace+"_connected_user",
			"connected users list",
			[]string{"username"},
			nil,
		),
		descUserResources: prometheus.NewDesc(
			namespace+"_user_resources",
			"user clients",
			[]string{"username", "resource"},
			nil,
		),
		descProcesses: prometheus.NewDesc(
			namespace+"_processes",
			"",
			nil,
			nil,
		),
		descUptime: prometheus.NewDesc(
			namespace+"_uptime",
			"",
			nil,
			nil,
		),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.descRegisteredUser
	ch <- c.descConnectedUser
	ch <- c.descUserResources
	ch <- c.descProcesses
	ch <- c.descUptime
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var (
		sendedRegistered, sendedConnected []string
	)

	registeredUsers, err := c.client.RegisteredUsers()
	if err != nil {
		ch <- prometheus.NewInvalidMetric(c.descRegisteredUser, err)
		fmt.Println(err)
		return
	}
	for _, username := range registeredUsers {
		if !str.InSlice(username, sendedRegistered) {
			ch <- prometheus.MustNewConstMetric(
				c.descRegisteredUser,
				prometheus.GaugeValue,
				float64(1),
				username,
			)
			sendedRegistered = append(sendedRegistered, username)
		}
		if !str.InSlice(username, c.knownUsers) {
			c.knownUsers = append(c.knownUsers, username)
		}
	}
	for _, username := range str.Diff(registeredUsers, c.knownUsers) {
		if !str.InSlice(username, sendedRegistered) {
			ch <- prometheus.MustNewConstMetric(
				c.descRegisteredUser,
				prometheus.GaugeValue,
				float64(0),
				username,
			)
		}
	}

	connectedUsers, err := c.client.ConnectedUsers()
	if err != nil {
		ch <- prometheus.NewInvalidMetric(c.descConnectedUser, err)
		fmt.Println(err)
	}
	for _, username := range connectedUsers {
		if !str.InSlice(username, sendedConnected) {
			ch <- prometheus.MustNewConstMetric(
				c.descConnectedUser,
				prometheus.GaugeValue,
				float64(1),
				username,
			)
			sendedConnected = append(sendedConnected, username)
		}

		if !str.InSlice(username, c.knownUsers) {
			c.knownUsers = append(c.knownUsers, username)
		}
	}
	for _, username := range str.Diff(connectedUsers, c.knownUsers) {
		if !str.InSlice(username, sendedConnected) {
			ch <- prometheus.MustNewConstMetric(
				c.descConnectedUser,
				prometheus.GaugeValue,
				float64(0),
				username,
			)
		}
	}

	for _, username := range connectedUsers {
		resources, err := c.client.UserResources(username)
		if err != nil {
			ch <- prometheus.NewInvalidMetric(c.descUserResources, err)
			fmt.Println(err)
			return
		}

		for _, device := range resources {
			ch <- prometheus.MustNewConstMetric(c.descUserResources, prometheus.GaugeValue, float64(1), username, device)
		}
	}

	uptime, err := c.client.Stats("uptimeseconds")
	if err != nil {
		ch <- prometheus.NewInvalidMetric(c.descUptime, err)
		fmt.Println(err)
	}
	ch <- prometheus.MustNewConstMetric(c.descUptime, prometheus.GaugeValue, float64(uptime))

	processes, err := c.client.Stats("processes")
	if err != nil {
		ch <- prometheus.NewInvalidMetric(c.descProcesses, err)
		fmt.Println(err)
	}
	ch <- prometheus.MustNewConstMetric(c.descProcesses, prometheus.GaugeValue, float64(processes))
}
