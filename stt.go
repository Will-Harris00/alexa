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

func CheckErr(w http.ResponseWriter, e error, errResp int) {
	if e != nil {
		println(e.Error())
		// println(errResp)
		w.WriteHeader(errResp) // tells our microservice what error response code to return
	}
}

func SpeechDecoding(w http.ResponseWriter, r *http.Request) []byte {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	CheckErr(w, err, http.StatusBadRequest) // could not decode json response due to perceived client error

	questionSpeech, ok := t["speech"].(string)

	if !ok { // speech field is not present
		err = errors.New("Object contains no field 'speech'") // handle error for incorrect json object
		CheckErr(w, err, http.StatusBadRequest)
	}

	// println(len(speech)) # check the entire base64 string is read
	if len(questionSpeech) < 5 || questionSpeech[0:5] != "UklGR" { // all wav files start with "UklGR" in base 64 standard encoding
		err = errors.New("Not a valid wav audio encoding!") // the audio file is invalid
		CheckErr(w, err, http.StatusBadRequest)
	}

	// DecodeString takes a base64 encoded string and returns the decoded data as a byte slice.
	// It will also return an error in case the input string has invalid base64 data.
	// StdEncoding: standard base64 encoding
	decodedSpeech, err := base64.StdEncoding.DecodeString(questionSpeech)
	CheckErr(w, err, http.StatusBadRequest) // could not encode speech due to malformed input perceived to be client error
	return decodedSpeech
}

func SpeechToText(w http.ResponseWriter, decodedSpeech []byte) []byte {
	client := &http.Client{}
	sttReq, err := http.NewRequest("POST", URI, bytes.NewReader(decodedSpeech))
	CheckErr(w, err, http.StatusBadRequest) // the request was malformed

	sttReq.Header.Set("Content-Type", "audio/wav;codecs=audio/pcm;samplerate=16000")
	sttReq.Header.Set("Ocp-Apim-Subscription-Key", KEY)

	sttResp, err := client.Do(sttReq)
	CheckErr(w, err, http.StatusBadRequest) // the server did not understand the request

	// the request was not successful
	if sttResp.StatusCode != http.StatusOK {
		CheckStatusErr(sttResp.StatusCode) // long text error message
		err = errors.New("Cannot convert speech to text!")
		CheckErr(w, err, sttResp.StatusCode) // pass the microsoft error code to our own microservice response header
	}

	defer sttResp.Body.Close() // defer ensures the response body is closed even in case of runtime error during parsing of response

	responseText, err := ioutil.ReadAll(sttResp.Body)
	CheckErr(w, err, http.StatusBadRequest) // the server cannot process the request due to something that is perceived to be a client error
	return responseText
}

func CheckResponse(w http.ResponseWriter, responseText []byte) string {
	t := map[string]interface{}{}
	err := json.Unmarshal(responseText, &t)
	CheckErr(w, err, http.StatusBadRequest) // could not decode json response due to perceived client error

	recStatus, ok := t["RecognitionStatus"].(string)
	if !ok { // RecognitionStatus field is not present
		err = errors.New("Object contains no field 'RecognitionStatus'") // handle error for incorrect json object
		CheckErr(w, err, http.StatusBadRequest)
	}

	if recStatus != "Success" { // recognition was not successful
		RecognitionErr(recStatus)
		err = errors.New("Text could not be determined!") // microsoft stt api failed to determine the correct text
		CheckErr(w, err, http.StatusBadRequest)
	}

	questionText, ok := t["DisplayText"].(string)
	if !ok { // DisplayText field is not present.
		err = errors.New("Object contains no field 'DisplayText'") // handle error for incorrect json object
		CheckErr(w, err, http.StatusBadRequest)
	}
	println(questionText)
	return questionText
}

func CheckStatusErr(errStatus int) {
	// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-speech-to-text
	// error handling for each status code
	// println(rsp.StatusCode)
	if errStatus == http.StatusBadRequest { // 400
		println("Bad request - The language code wasn't provided, the language isn't supported, " +
			"or the audio file is invalid (for example).")
	}
	if errStatus == http.StatusUnauthorized { // 401
		println("Unauthorized - A subscription key or an authorization token is invalid in the specified region, " +
			"or an endpoint is invalid.")
	}
	if errStatus == http.StatusForbidden { // 403
		println("Forbidden - A subscription key or authorization token is missing.")
	}
}

func RecognitionErr(recStatus string) {
	// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-speech-to-text#pronunciation-assessment-parameters
	// Determines the error type from the response parameters for RecognitionStatus
	switch {
	case recStatus == "NoMatch":
		println("Speech was detected in the audio stream, but no words from the target language were matched. " +
			"This status usually means that the recognition language is different from the language that the user is speaking.")
	case recStatus == "InitialSilenceTimeout":
		println("The start of the audio stream contained only silence, and the service timed out while waiting for speech.")
	case recStatus == "BabbleTimeout":
		println("The start of the audio stream contained only noise, and the service timed out while waiting for speech.")
	case recStatus == "Error":
		println("The recognition service encountered an internal error and could not continue. Try again if possible.")
	}
}

func STTResponse(w http.ResponseWriter, questionText string) {
	u := map[string]interface{}{"text": questionText}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func ProcessSTT(w http.ResponseWriter, r *http.Request) {
	decodedSpeech := SpeechDecoding(w, r)
	responseText := SpeechToText(w, decodedSpeech)
	questionText := CheckResponse(w, responseText)
	STTResponse(w, questionText)
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
