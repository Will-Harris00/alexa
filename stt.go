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

func SpeechToText(speech []byte) (string, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("POST", URI, bytes.NewReader(speech))

	req.Header.Set("Content-Type",
		"audio/wav;codecs=audio/pcm;samplerate=16000;base64")
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)

	rsp, _ := client.Do(req)

	defer rsp.Body.Close()

	if rsp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(rsp.Body)
		return string(body), nil
	} else {
		return "", errors.New("Cannot convert to speech to text!")
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
			text, _ := SpeechToText(bytes_slice)
			println(text)
		}
	}
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
