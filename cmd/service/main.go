package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type detail struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Fatal  bool   `json:"fatal"`
	Error  string `json:"error,omitempty"`
}
type response struct {
	Status    string   `json:"status"`
	Timestamp string   `json:"timestamp"`
	Details   []detail `json:"details"`
}

func main() {
	address := flag.String("address", ":1234", "address to listen to")
	failedComponents := flag.String("failedComponents", "service1,service2", "comma separated list of the failed components")
	okComponents := flag.String("okComponents", "service3,service4", "comma separated list of the ok components")
	flag.Parse()

	details := make([]detail, 0)
	for _, comp := range strings.Split(*failedComponents, ",") {
		fatal := false
		if rand.Intn(10) > 5 {
			fatal = true
		}

		details = append(details, detail{
			Name:   comp,
			Status: "failed",
			Fatal:  fatal,
			Error:  "some error happened",
		})

	}

	for _, comp := range strings.Split(*okComponents, ",") {
		fatal := false
		if rand.Intn(10) > 5 {
			fatal = true
		}
		details = append(details, detail{
			Name:   comp,
			Status: "ok",
			Fatal:  fatal,
		})

	}

	resp := response{Status: "ok", Timestamp: time.Now().String(), Details: details}

	http.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		bts, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bts)
	})

	log.Printf("Listen on %s\n", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
