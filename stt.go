package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".stt.speech.microsoft.com/" +
		"speech/recognition/conversation/cognitiveservices/v1?" +
		"language=en-US"
	KEY = ""
)

func CheckErr(w http.ResponseWriter, e error, err_resp int) {
	if e != nil {
		println(e.Error())
		w.WriteHeader(err_resp) // tells our microservice what error response code to return
	}
}

func SpeechDecoding(w http.ResponseWriter, r *http.Request) []byte {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	CheckErr(w, err, http.StatusBadRequest) // could not decode json response due to perceived client error

	question_speech, ok := t["speech"].(string)

	if !ok { // speech field is not present
		err = errors.New("Object contains no field 'speech'") // handle error for incorrect json object
		CheckErr(w, err, http.StatusBadRequest)
	}

	// println(len(speech)) # check the entire base64 string is read
	if len(question_speech) < 5 || question_speech[0:5] != "UklGR" { // all wav files start with "UklGR" in base 64 standard encoding
		err = errors.New("Not a valid wav audio encoding!") // the audio file is invalid
		CheckErr(w, err, http.StatusBadRequest)
	}

	// DecodeString takes a base64 encoded string and returns the decoded data as a byte slice.
	// It will also return an error in case the input string has invalid base64 data.
	// StdEncoding: standard base64 encoding
	decoded_speech, err := base64.StdEncoding.DecodeString(question_speech)
	CheckErr(w, err, http.StatusBadRequest) // could not encode speech due to malformed input perceived to be client error
	return decoded_speech
}

func SpeechToText(w http.ResponseWriter, speech []byte) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewReader(speech))
	CheckErr(w, err, http.StatusBadRequest) // the request was malformed

	req.Header.Set("Content-Type", "audio/wav;codecs=audio/pcm;samplerate=16000")
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)

	rsp, err := client.Do(req)
	CheckErr(w, err, http.StatusBadRequest) // the server did not understand the request

	defer rsp.Body.Close() // defer ensures the response body is closed even in case of runtime error during parsing of response

	// the request was not successful
	if rsp.StatusCode != http.StatusOK {
		CheckStatusErr(rsp.StatusCode) // long text error message
		err = errors.New("Cannot convert speech to text!")
		CheckErr(w, err, rsp.StatusCode) // pass the microsoft error code to our own microservice response header
	}

	response_text, err := ioutil.ReadAll(rsp.Body)
	CheckErr(w, err, http.StatusBadRequest) // the server cannot process the request due to something that is perceived to be a client error
	return response_text
}

func CheckResponse(w http.ResponseWriter, body []byte) string {
	t := map[string]interface{}{}
	err := json.Unmarshal(body, &t)
	CheckErr(w, err, http.StatusBadRequest) // could not decode json response due to perceived client error

	rec_status, ok := t["RecognitionStatus"].(string)
	if !ok { // RecognitionStatus field is not present
		err = errors.New("Object contains no field 'RecognitionStatus'") // handle error for incorrect json object
		CheckErr(w, err, http.StatusBadRequest)
	}

	if rec_status != "Success" { // recognition was not successful
		RecognitionErr(rec_status)
		err = errors.New("Text could not be determined!") // microsoft stt api failed to determine the correct text
		CheckErr(w, err, http.StatusBadRequest)
	}

	question_text, ok := t["DisplayText"].(string)
	if !ok { // DisplayText field is not present.
		err = errors.New("Object contains no field 'DisplayText'") // handle error for incorrect json object
		CheckErr(w, err, http.StatusBadRequest)
	}
	println(question_text)
	return question_text
}

func CheckStatusErr(err_status int) {
	// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-speech-to-text
	// error handling for each status code
	// println(rsp.StatusCode)
	if err_status == http.StatusBadRequest { // 400
		println("Bad request - The language code wasn't provided, the language isn't supported, " +
			"or the audio file is invalid (for example).")
	}
	if err_status == http.StatusUnauthorized { // 401
		println("Unauthorized - A subscription key or an authorization token is invalid in the specified region, " +
			"or an endpoint is invalid.")
	}
	if err_status == http.StatusForbidden { // 403
		println("Forbidden - A subscription key or authorization token is missing.")
	}
}

func RecognitionErr(rec_status string) {
	// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-speech-to-text#pronunciation-assessment-parameters
	// Determines the error type from the response parameters for RecognitionStatus
	switch {
	case rec_status == "NoMatch":
		println("Speech was detected in the audio stream, but no words from the target language were matched. " +
			"This status usually means that the recognition language is different from the language that the user is speaking.")
	case rec_status == "InitialSilenceTimeout":
		println("The start of the audio stream contained only silence, and the service timed out while waiting for speech.")
	case rec_status == "BabbleTimeout":
		println("The start of the audio stream contained only noise, and the service timed out while waiting for speech.")
	case rec_status == "Error":
		println("The recognition service encountered an internal error and could not continue. Try again if possible.")
	}
}

func STTResponse(w http.ResponseWriter, question_text string) {
	u := map[string]interface{}{"text": question_text}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func ProcessSTT(w http.ResponseWriter, r *http.Request) {
	decoded_speech := SpeechDecoding(w, r)
	response_text := SpeechToText(w, decoded_speech)
	question_text := CheckResponse(w, response_text)
	STTResponse(w, question_text)
}

func STTHandler() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/stt", ProcessSTT).Methods("POST")
	http.ListenAndServe(":3002", r)
	//	3001 / alpha
	//	3002 / stt
	//	3003 / tts
}

func main() {
	STTHandler()
}
