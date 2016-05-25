package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// incHandler is a HTTP Handler for increment requets. Takes the form of /inc?amount=1
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

	fmt.Printf("Incremented counter to %v\n", val)
	w.Write([]byte(val))

}

// getHandler is a HTTP Handler to fetch the counter's count. Just /
func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	val := strconv.Itoa(counter.Count())
	w.Write([]byte(val))
}

// HTTP Handler to fetch the cluster membership state
func clusterHandler(w http.ResponseWriter, r *http.Request) {

	json.NewEncoder(w).Encode(m.Members())

}
