package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// HTTP Handler for increment requests. Takes the form of /inc?amount=1
func incHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	amountForm := r.Form.Get("amount")

	//parse inc amount
	amount, parseErr := strconv.Atoi(amountForm)

	if parseErr != nil {
		http.Error(w, parseErr.Error(), 500)
		return
	}

	if amount < 1 {
		http.Error(w, "Only positive amounts are supported", 501)
		return
	}

	counter.IncVal(amount)

	fmt.Fprintln(w, counter)

	// broadcast the state?
	// go BroadcastState()
}

// HTTP Handler to fetch the counter's count. Just /
func getHandler(w http.ResponseWriter, r *http.Request) {

	val := strconv.Itoa(counter.Count())

	fmt.Fprintln(w, counter)
}

// HTTP Handler to fetch the full local CRDT's counter state
func verboseHandler(w http.ResponseWriter, r *http.Request) {

	counterJSON, marshalErr := counter.MarshalJSON()

	if marshalErr != nil {
		http.Error(w, marshalErr.Error(), 500)
		return
	}

	w.Write(counterJSON)
}

// HTTP Handler to fetch the cluster membership state
func clusterHandler(w http.ResponseWriter, r *http.Request) {

	json.NewEncoder(w).Encode(m.Members())

}
