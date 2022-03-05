package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	URI = "http://api.wolframalpha.com/v1/result"
	KEY = ""
)

func ProcessAlpha(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode((&t))
	if err != nil {
		AlphaErrResponse(w, err, http.StatusBadRequest) // bad request due to perceived client error
	} else {
		textQuery := t["text"].(string)

		alphaResp, wolframStatus, err, errCode := AlphaService(textQuery)
		if err != nil {
			AlphaErrResponse(w, err, errCode) // return an error response from the microservice
		} else {
			AlphaResponse(w, alphaResp, wolframStatus) // success
		}
	}
}

func AlphaService(textQuery string) ([]byte, int, error, int) {
	println(textQuery)                                                     // check the question
	alphaURI := URI + "?appid=" + KEY + "&i=" + url.QueryEscape(textQuery) // html encoded string

	wolframResp, err := http.Get(alphaURI)
	if err != nil {
		return nil, 0, err, http.StatusBadRequest // the request was malformed
	}

	// println(wolframResponse.StatusCode) // determine if the wolfram alpha api returned the correct response
	err = CheckStatusError(wolframResp.StatusCode)
	if err != nil {
		return nil, 0, err, wolframResp.StatusCode // copy the status code returned from wolfram alpha short answers api
	}

	defer wolframResp.Body.Close() // delay the execution of the function until the nearby functions returns

	wolframRespBody, err := ioutil.ReadAll(wolframResp.Body) // read the body of the response returned from the wolfram api
	if err != nil {
		return nil, 0, err, http.StatusInternalServerError // could not read the body of the response, perceived to be client error
	}

	println(string(wolframRespBody))

	return wolframRespBody, wolframResp.StatusCode, nil, 0
}

func CheckStatusError(status int) error {
	switch status {
	case http.StatusBadRequest: // 400 - No input.  Please specify the input using the 'i' query parameter.
		err := errors.New("This status indicates that the API did not find an input parameter while parsing. " +
			"In most cases, this can be fixed by checking that you have used the correct syntax for " +
			"including the i parameter.")
		return err
	case http.StatusForbidden: // 403 - Error 1: Invalid appid or Error 2: Appid missing
		err := errors.New("This error is returned when a request contains an invalid option for the appid parameter. " +
			"Double-check that your AppID is typed correctly and that your appid parameter is using the correct syntax.")
		return err
	case http.StatusNotImplemented: // 501 - No short answer available
		err := errors.New("This status is returned if a given input value cannot be interpreted by this API. " +
			"This is commonly caused by input that is misspelled, poorly formatted or otherwise unintelligible. " +
			"Because this API is designed to return a single result, this message may appear if no sufficiently " +
			"short result can be found. You may occasionally receive this status when requesting information on " +
			"topics that are restricted or not covered.")
		return err
	}
	return nil
}

func AlphaResponse(w http.ResponseWriter, alphaResp []byte, wolframStatus int) {
	w.WriteHeader(http.StatusOK)
	u := map[string]interface{}{"text": string(alphaResp)}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	json.NewEncoder(w).Encode(u)                       // encode string text as json object
}

func AlphaErrResponse(w http.ResponseWriter, err error, errCode int) {
	w.WriteHeader(errCode)
	w.Write([]byte(err.Error()))
	w.Header().Set("Content-Type", "text") // return error message as text
	println(errCode)
	println(err.Error()) // display the error message on the console
}

func main() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alpha", ProcessAlpha).Methods("POST")
	http.ListenAndServe(":3001", r)
}
