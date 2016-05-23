package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/memberlist"
	"github.com/nphase/crdt"

	uuid "github.com/satori/go.uuid"
)

var (
	counter = &crdt.GCounter{}

	members = flag.String("members", "", "comma seperated list of members")
	port    = flag.Int("port", 4001, "http port")

	broadcasts *memberlist.TransmitLimitedQueue
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type delegate struct{}

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

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

//Handle merge events via gossip
func (d *delegate) NotifyMsg(b []byte) {

	if len(b) == 0 {
		return
	}

	switch b[0] {
	case 'd': // data
		var update *update
		if err := json.Unmarshal(b[1:], &update); err != nil {
			return
		}

		switch update.Action {
		case "merge":
			external_crdt := crdt.NewGCounterFromJSONBytes([]byte(update.Data))
			counter.Merge(external_crdt)
		}

	}
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {

	b, _ := counter.MarshalJSON()

	return b
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}

	external_crdt := crdt.NewGCounterFromJSONBytes(buf)
	counter.Merge(external_crdt)

}

//handle inc Request
func incHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	amount_form_value := r.Form.Get("amount")

	//parse inc amount
	amount, parse_err := strconv.Atoi(amount_form_value)

	if parse_err != nil {
		http.Error(w, parse_err.Error(), 500)
		return
	}

	if amount < 0 {
		http.Error(w, "Deprecation not supported", 501)
		return
	}

	counter.IncVal(amount)

	val := strconv.Itoa(counter.Count())
	w.Write([]byte(val))

	//broadcast the state
	go BroadcastState()
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	val := strconv.Itoa(counter.Count())

	w.Write([]byte(val))
}

func verboseHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	counter_json, marshal_err := counter.MarshalJSON()

	if marshal_err != nil {
		http.Error(w, marshal_err.Error(), 500)
		return
	}

	w.Write(counter_json)
}

//Broadcast a message to all members
func BroadcastState() {

	counter_json, marshal_err := counter.MarshalJSON()

	if marshal_err != nil {
		panic("Failed to marshal counter state in BroadcastState()")
		return
	}

	b, err := json.Marshal(&update{
		Action: "merge",
		Data:   counter_json,
	})

	if err != nil {
		panic("Failed to marshal broadcast message in BroadcastState()")
		return
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})

}

func start() error {
	flag.Parse()

	counter = crdt.NewGCounter()

	hostname, _ := os.Hostname()
	c := memberlist.DefaultWANConfig()
	c.Delegate = &delegate{}
	c.BindPort = 0
	c.Name = hostname + "-" + uuid.NewV4().String()
	m, err := memberlist.Create(c)
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

	http.HandleFunc("/verbose", verboseHandler)
	http.HandleFunc("/inc", incHandler)

	http.HandleFunc("/", getHandler)

	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}
