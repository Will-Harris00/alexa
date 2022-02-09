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

func SpeechToText(speech []byte) ([]byte, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("POST", URI, bytes.NewReader(speech))

	req.Header.Set("Content-Type",
		"audio/wav;codecs=audio/pcm;samplerate=16000;base64")
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)

	rsp, _ := client.Do(req)

	defer rsp.Body.Close()

	println(rsp.StatusCode)

	if rsp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(rsp.Body)
		return body, nil
	} else {
		return nil, errors.New("Cannot convert to speech to text!")
	}
}

func Speech(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if speech, ok := t["speech"].(string); ok {
			// println(len(speech)) # check the entire base64 string is read
			// DecodeString takes a base64 encoded string and returns the decoded data as a byte slice.
			// It will also return an error in case the input string has invalid base64 data.
			// StdEncoding: standard base64 encoding
			bytes_slice, err := base64.StdEncoding.DecodeString(speech)
			if err != nil {
				println("Malformed input!")
			}
			body, _ := SpeechToText(bytes_slice)
			// println(text)
			STTResponse(w, body)
		}
	}
}

func CheckReponse(body []byte) string {
	t := map[string]interface{}{}
	err := json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}

	if rec_status, ok := t["RecognitionStatus"].(string); ok {
		if rec_status == "Success" { // recognition was successful, and the DisplayText field is present.
			if plain_text, ok := t["DisplayText"].(string); ok {
				print(plain_text)
				return plain_text
			}
		} else {
			RecognitionStatus(rec_status)
			panic("Text could not be determined!") // api failed to determine the correct text
		}
	} else {
		panic("Object contains no field 'RecognitionStatus'") // handle error for incorrect json object
	}
	panic("Object does not contain 'DisplayText'") // failed to return text
}

func RecognitionStatus(rec_status string) {
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

func STTResponse(w http.ResponseWriter, body []byte) {
	plain_text := CheckReponse(body)
	u := map[string]interface{}{"text": plain_text}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func main() {
	STTHandler()
}

func STTHandler() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/stt", Speech).Methods("POST")
	http.ListenAndServe(":3002", r)
	//	3001 / alpha
	//	3002 / stt
	//	3003 / tts
}
