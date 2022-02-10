package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
)

func Alexa(w http.ResponseWriter, r *http.Request) {
	stt_rsp := SpeechToTextManager(r)
	alpha_rsp := AlphaManager(stt_rsp)
	tts_rsp_body := TextToSpeechManager(alpha_rsp)
	AlexaResponse(w, tts_rsp_body)
}

func SpeechToTextManager(r *http.Request) []byte {
	stt_uri := "http://localhost:3002/stt"

	req, _ := http.NewRequest("POST", stt_uri, r.Body)
	req.Header.Set("Content-Type", "application/json")

	rsp, _ := http.DefaultClient.Do(req)

	if rsp.StatusCode != http.StatusOK {
		panic("Something went wrong with the speech-to-text microservice!")
	}

	defer rsp.Body.Close()

	if rsp.StatusCode == http.StatusOK {
		rsp_body, _ := ioutil.ReadAll(rsp.Body)
		// println(string(rsp_body)) // check response is received from speech-to-text microservice
		return rsp_body
	}
	return nil
}

func AlphaManager(stt_rsp []byte) io.ReadCloser {
	alpha_uri := "http://localhost:3001/alpha"

	req, _ := http.NewRequest("POST", alpha_uri, bytes.NewReader(stt_rsp))
	req.Header.Set("Content-Type", "application/json")

	alpha_rsp, _ := http.DefaultClient.Do(req)

	if alpha_rsp.StatusCode != http.StatusOK {
		panic("Something went wrong with the alpha microservice!")
	}

	return alpha_rsp.Body
}

func TextToSpeechManager(alpha_rsp io.ReadCloser) []byte {
	tts_uri := "http://localhost:3003/tts"

	req, _ := http.NewRequest("POST", tts_uri, alpha_rsp)
	req.Header.Set("Content-Type", "application/json")

	tts_rsp, _ := http.DefaultClient.Do(req)

	if tts_rsp.StatusCode != http.StatusOK {
		panic("Something went wrong with the text-to-speech microservice!")
	}

	defer tts_rsp.Body.Close()

	if tts_rsp.StatusCode == http.StatusOK {
		rsp_body, _ := ioutil.ReadAll(tts_rsp.Body)
		// println(string(rsp_body)) // check response is received from speech-to-text microservice
		return rsp_body
	}
	return nil
}

func AlexaResponse(w http.ResponseWriter, tts_rsp_body []byte) {
	t := map[string]interface{}{}
	err := json.Unmarshal(tts_rsp_body, &t)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)
	println("Play the answer.wav file to hear the solution to your question!")
}

func AlexaHandler() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alexa", Alexa).Methods("POST")
	http.ListenAndServe(":3000", r)
	//	3001 / alpha
	//	3002 / stt
	//	3003 / tts
}

func main() {
	AlexaHandler()
}
