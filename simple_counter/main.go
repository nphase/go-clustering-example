package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
)

var (
	port    = flag.Int("port", 4000, "http port")
	counter *Counter
)

// START OMIT
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

func (c *Counter) String() string {
	return strconv.Itoa(counter.Count())
}

// END OMIT

// incHandler is a HTTP Handler for increment requests. Takes the form of /inc?amount=1
func incHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	//parse inc amount
	amount, parseErr := strconv.Atoi(r.FormValue("amount"))

	if parseErr != nil {
		http.Error(w, parseErr.Error(), 500)
		return
	}

	if amount < 0 {
		http.Error(w, "Deprecation not supported", 501)
		return
	}

	counter.IncVal(amount)

	fmt.Printf("Incremented counter to %v\n", counter)
	fmt.Fprintln(w, counter)

}

// getHandler is a HTTP Handler to fetch the counter's count. Just /
func getHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, counter)
}

func main() {
	flag.Parse()

	counter = &Counter{}

	http.HandleFunc("/inc", incHandler)
	http.HandleFunc("/", getHandler)

	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}
