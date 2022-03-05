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

func ProcessSTT(w http.ResponseWriter, r *http.Request) {
	decodedSpeech, err, errCode := SpeechDecoding(r)
	if err != nil {
		STTErrResponse(w, err, errCode) // return an error response from the microservice
	}

	responseText, err, errCode := SpeechToText(decodedSpeech)
	if err != nil {
		STTErrResponse(w, err, errCode) // return an error response from the microservice
	}

	questionText, err, errCode := CheckResponse(responseText)
	if err != nil {
		STTErrResponse(w, err, errCode) // return an error response from the microservice
	} else {
		STTResponse(w, questionText) // success
	}
}

func SpeechDecoding(r *http.Request) ([]byte, error, int) {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return nil, err, http.StatusBadRequest // could not decode json query due to perceived client error
	}

	questionSpeech, ok := t["speech"].(string)

	if !ok { // speech field is not present
		err = errors.New("Object contains no field 'speech'") // handle error for incorrect json object
		return nil, err, http.StatusBadRequest
	}

	// println(len(speech)) # check the entire base64 string is read
	if len(questionSpeech) < 5 || questionSpeech[0:5] != "UklGR" { // all wav files start with "UklGR" in base 64 standard encoding
		err = errors.New("Not a valid wav audio encoding!") // the audio file is invalid
		return nil, err, http.StatusBadRequest
	}

	// DecodeString takes a base64 encoded string and returns the decoded data as a byte slice.
	// It will also return an error in case the input string has invalid base64 data.
	// StdEncoding: standard base64 encoding
	decodedSpeech, err := base64.StdEncoding.DecodeString(questionSpeech)
	if err != nil {
		return nil, err, http.StatusBadRequest // could not encode speech due to malformed input, perceived to be client error
	}

	return decodedSpeech, nil, 0
}

func SpeechToText(decodedSpeech []byte) ([]byte, error, int) {
	client := &http.Client{}
	sttReq, err := http.NewRequest("POST", URI, bytes.NewReader(decodedSpeech))
	if err != nil {
		return nil, err, http.StatusBadRequest // the request was malformed
	}

	sttReq.Header.Set("Content-Type", "audio/wav;codecs=audio/pcm;samplerate=16000")
	sttReq.Header.Set("Ocp-Apim-Subscription-Key", KEY)

	sttResp, err := client.Do(sttReq)
	if err != nil {
		return nil, err, http.StatusNotFound // microsoft speech-to-text could not be reached
	}

	// the request was not successful
	if sttResp.StatusCode != http.StatusOK {
		err = CheckSTTStatusErr(sttResp.StatusCode) // long text error message
		if err != nil {
			return nil, err, sttResp.StatusCode // pass the microsoft stt error code to our own microservice response header
		}
	}

	defer sttResp.Body.Close() // defer ensures the response body is closed even in case of runtime error during parsing of response

	responseText, err := ioutil.ReadAll(sttResp.Body)
	if err != nil {
		return nil, err, http.StatusInternalServerError // could not read the body of the response, perceived to be client error
	}

	println(string(responseText))

	return responseText, nil, 0
}

func CheckResponse(responseText []byte) (string, error, int) {
	t := map[string]interface{}{}
	err := json.Unmarshal(responseText, &t)
	if err != nil {
		return "", err, http.StatusBadRequest // could not decode json response due to perceived client error
	}

	recStatus, ok := t["RecognitionStatus"].(string)
	if !ok { // RecognitionStatus field is not present
		err = errors.New("Object contains no field 'RecognitionStatus'") // handle error for incorrect json object
		return "", err, http.StatusInternalServerError
	}

	if recStatus != "Success" { // recognition was not successful
		err = RecognitionErr(recStatus) // microsoft stt api failed to determine the correct text
		return "", err, http.StatusInternalServerError
	}

	questionText, ok := t["DisplayText"].(string)
	if !ok { // DisplayText field is not present.
		err = errors.New("Object contains no field 'DisplayText'") // handle error for incorrect json object
		return "", err, http.StatusInternalServerError
	}

	println(questionText)

	return questionText, nil, 0
}

func CheckSTTStatusErr(errStatus int) error {
	// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-speech-to-text
	// error handling for each status code
	// println(errStatus)
	switch errStatus {
	case http.StatusBadRequest: // 400
		err := errors.New("Bad request - The language code wasn't provided, the language isn't supported, " +
			"or the audio file is invalid (for example).")
		return err
	case http.StatusUnauthorized: // 401
		err := errors.New("Unauthorized - A subscription key or an authorization token is invalid in the specified region, " +
			"or an endpoint is invalid.")
		return err
	case http.StatusForbidden: // 403
		err := errors.New("Forbidden - A subscription key or authorization token is missing.")
		return err
	}
	return nil
}

func RecognitionErr(recStatus string) error {
	// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-speech-to-text#pronunciation-assessment-parameters
	// Determines the error type from the response parameters for RecognitionStatus
	switch {
	case recStatus == "NoMatch":
		err := errors.New("Speech was detected in the audio stream, but no words from the target language were matched. " +
			"This status usually means that the recognition language is different from the language that the user is speaking.")
		return err
	case recStatus == "InitialSilenceTimeout":
		err := errors.New("The start of the audio stream contained only silence, and the service timed out while waiting for speech.")
		return err
	case recStatus == "BabbleTimeout":
		err := errors.New("The start of the audio stream contained only noise, and the service timed out while waiting for speech.")
		return err
	case recStatus == "Error":
		err := errors.New("The recognition service encountered an internal error and could not continue. Try again if possible.")
		return err
	}
	return nil
}

func STTResponse(w http.ResponseWriter, questionText string) {
	u := map[string]interface{}{"text": questionText}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func STTErrResponse(w http.ResponseWriter, err error, errCode int) {
	w.WriteHeader(errCode)
	w.Write([]byte(err.Error()))
	w.Header().Set("Content-Type", "text") // return error message as text
	println(errCode)
	println(err.Error()) // display the error message on the console
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
