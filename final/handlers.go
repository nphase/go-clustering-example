package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// HTTP Handler for increment requets. Takes the form of /inc?amount=1
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

// HTTP Handler to fetch the counter's count. Just /
func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	val := strconv.Itoa(counter.Count())

	w.Write([]byte(val))
}

// HTTP Handler to fetch the full local CRDT's counter state
func verboseHandler(w http.ResponseWriter, r *http.Request) {

	counter_json, marshal_err := counter.MarshalJSON()

	if marshal_err != nil {
		http.Error(w, marshal_err.Error(), 500)
		return
	}

	w.Write(counter_json)
}

// HTTP Handler to fetch the cluster membership state
func clusterHandler(w http.ResponseWriter, r *http.Request) {

	json.NewEncoder(w).Encode(m.Members())

}
