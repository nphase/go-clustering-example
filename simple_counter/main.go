package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
)

var (
	port    = flag.Int("port", 4001, "http port")
	counter *Counter
)

type Counter struct {
	val int32
}

// IncVal increments the counter's value by d
func (c *Counter) IncVal(d int) {

	atomic.AddInt32(&c.val, int32(d))

}

// Count fetches the counter value
func (c *Counter) Count() int {

	return int(atomic.LoadInt32(&c.val))

}

//handle inc Request
func incHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	amountForm := r.Form.Get("amount")

	//parse inc amount
	amount, parseErr := strconv.Atoi(amountForm)

	if parseErr != nil {
		http.Error(w, parseErr.Error(), 500)
		return
	}

	if amount < 0 {
		http.Error(w, "Deprecation not supported", 501)
		return
	}

	counter.IncVal(amount)

	val := strconv.Itoa(counter.Count())
	w.Write([]byte(val))

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	val := strconv.Itoa(counter.Count())
	w.Write([]byte(val))
}

func main() {
	flag.Parse()

	http.HandleFunc("/inc", incHandler)
	http.HandleFunc("/", getHandler)

	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}
