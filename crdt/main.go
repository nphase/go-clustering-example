package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/nphase/crdt"

	uuid "github.com/satori/go.uuid"
)

var (
	counter = &crdt.GCounter{}

	members  = flag.String("members", "", "comma seperated list of members")
	port     = flag.Int("port", 4001, "http port")
	rpc_port = flag.Int("rpc_port", 0, "memberlist port (0 = auto select)")

	m *memberlist.Memberlist
)

func start() error {
	flag.Parse()

	counter = crdt.NewGCounter()

	hostname, _ := os.Hostname()
	c := memberlist.DefaultWANConfig()
	c.BindPort = *rpc_port
	c.Name = hostname + "-" + uuid.NewV4().String()

	c.PushPullInterval = time.Second * 3 // to make sync demonstrable
	c.ProbeInterval = time.Second * 1    // to make failure demonstrable

	var err error

	m, err = memberlist.Create(c)

	if err != nil {
		return err
	}

	if len(*members) > 0 {
		parts := strings.Split(*members, ",")
		_, err := m.Join(parts)
		if err != nil {
			return err
		}
	}

	node := m.LocalNode()
	fmt.Printf("Local member %s:%d\n", node.Addr, node.Port)

	return nil
}

func main() {
	if err := start(); err != nil {
		fmt.Println(err)
	}

	http.HandleFunc("/cluster", clusterHandler)
	http.HandleFunc("/verbose", verboseHandler)
	http.HandleFunc("/inc", incHandler)
	http.HandleFunc("/", getHandler)

	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}
