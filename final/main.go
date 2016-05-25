package main

import (
	"encoding/json"
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

	broadcasts *memberlist.TransmitLimitedQueue

	m *memberlist.Memberlist
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type update struct {
	Action string          // merge
	Data   json.RawMessage // crdt.GCounterJSON
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

type delegate struct{}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

//Handle merge events via gossip
func (d *delegate) NotifyMsg(b []byte) {

	if len(b) == 0 {
		return
	}

	var update *update
	if err := json.Unmarshal(b, &update); err != nil {
		return
	}

	switch update.Action {
	case "merge":
		externalCRDT := crdt.NewGCounterFromJSONBytes([]byte(update.Data))
		counter.Merge(externalCRDT)
	default:
		panic("unsupported update action")
	}

}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

// Share the local counter state via MemberList to another node
func (d *delegate) LocalState(join bool) []byte {

	b, err := counter.MarshalJSON()

	if err != nil {
		panic(err)
	}

	return b
}

// Merge in received counter state whenever
// join = false means this was received after a push/pull sync
// basically a full state refresh.
func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}

	externalCRDT := crdt.NewGCounterFromJSONBytes(buf)
	counter.Merge(externalCRDT)
}

// BroadcastState broadcasts the local counter state to all cluster members
func BroadcastState() {

	counterJSON, marshalErr := counter.MarshalJSON()

	if marshalErr != nil {
		panic("Failed to marshal counter state in BroadcastState()")
	}

	b, err := json.Marshal(&update{
		Action: "merge",
		Data:   counterJSON,
	})

	if err != nil {
		panic("Failed to marshal broadcast message in BroadcastState()")
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    b,
		notify: nil,
	})

}

func start() error {
	flag.Parse()

	counter = crdt.NewGCounter()

	hostname, _ := os.Hostname()
	c := memberlist.DefaultWANConfig()
	c.Delegate = &delegate{}
	c.BindPort = *rpc_port
	c.Name = hostname + "-" + uuid.NewV4().String()

	c.PushPullInterval = time.Second * 5 // to make it demonstrable
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

	broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
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
