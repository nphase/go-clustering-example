package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

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
	val := strconv.Itoa(counter.Count())
	fmt.Fprintln(w, counter)
}

// HTTP Handler to fetch the cluster membership state
func clusterHandler(w http.ResponseWriter, r *http.Request) {

	json.NewEncoder(w).Encode(m.Members())

}
