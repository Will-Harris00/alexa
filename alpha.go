package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	URI = "http://api.wolframalpha.com/v1/result"
	KEY = ""
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Alpha(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode((&t))
	check(err)

	textQuery := t["text"].(string)

	alphaResponse, wolframStatus := AlphaService(textQuery)

	u := map[string]interface{}{"text": alphaResponse}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(wolframStatus)                       // copy the status code returned from wolfram alpha short answers api
	json.NewEncoder(w).Encode(u)                       // encode string text as json object
}

func AlphaService(textQuery string) (interface{}, int) {
	println(textQuery)                                                     // check the question
	alphaURI := URI + "?appid=" + KEY + "&i=" + url.QueryEscape(textQuery) // html encoded string

	wolframResponse, err := http.Get(alphaURI)
	check(err)

	CheckStatusError(wolframResponse.StatusCode)
	// println(wolframResponse.StatusCode) // determine if the wolfram alpha api returned the correct response

	defer wolframResponse.Body.Close() // delay the execution of the function until the nearby functions returns

	responseData, err := ioutil.ReadAll(wolframResponse.Body) // read the body of the response returned from the wolfram api
	check(err)

	responseString := string(responseData)
	println(responseString)

	return responseString, wolframResponse.StatusCode
}

func CheckStatusError(status int) {
	switch status {
	case http.StatusBadRequest: // 400 - No input.  Please specify the input using the 'i' query parameter.
		println("This status indicates that the API did not find an input parameter while parsing. " +
			"In most cases, this can be fixed by checking that you have used the correct syntax for " +
			"including the i parameter.")
	case http.StatusForbidden: // 403 - Error 1: Invalid appid or Error 2: Appid missing
		println("This error is returned when a request contains an invalid option for the appid parameter. " +
			"Double-check that your AppID is typed correctly and that your appid parameter is using the correct syntax.")
	case http.StatusNotImplemented: // 501 - No short answer available
		println("This status is returned if a given input value cannot be interpreted by this API. " +
			"This is commonly caused by input that is misspelled, poorly formatted or otherwise unintelligible. " +
			"Because this API is designed to return a single result, this message may appear if no sufficiently " +
			"short result can be found. You may occasionally receive this status when requesting information on " +
			"topics that are restricted or not covered.")
	}
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alpha", Alpha).Methods("POST")
	http.ListenAndServe(":3001", r)
}
