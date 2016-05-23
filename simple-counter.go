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
	val *int32
}

func (c *Counter) IncVal(d int) {

	atomic.AddInt32(c.val, 1)

}

func (c *Counter) Count() int {

	x := int(atomic.LoadInt32(c.val))

	return x
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

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	val := strconv.Itoa(counter.Count())
	w.Write([]byte(val))
}

func start() error {

	counter_val := int32(0)

	counter = &Counter{
		val: &counter_val,
	}

	return nil
}

func main() {
	flag.Parse()

	if err := start(); err != nil {
		fmt.Println(err)
	}

	http.HandleFunc("/inc", incHandler)
	http.HandleFunc("/", getHandler)

	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}
