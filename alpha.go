package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	URI = "http://api.wolframalpha.com/v1/result"
	KEY = ""
)

func Alpha(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if text_query, ok := t["text"].(string); ok {
			if response, err := AlphaService(text_query); err == nil {
				u := map[string]interface{}{"text": response}
				w.Header().Set("Content-Type", "application/json") // return microservice response as json
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(u) // encode string text as json object
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func AlphaService(text_query string) (interface{}, error) {
	alpha_uri := URI + "?appid=" + KEY + "&i=" + url.QueryEscape(text_query) // html encoded string

	response, err := http.Get(alpha_uri)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close() // delay the execution of the function until the nearby functions returns

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)

	println(response.Body)
	println(responseData)
	println(responseString)

	return responseString, nil
	//return nil, errors.New("Service")
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alpha", Alpha).Methods("POST")
	http.ListenAndServe(":3001", r)
}
