package main

import (
	"bytes"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
)

func Alexa(w http.ResponseWriter, r *http.Request) {
	stt_rsp := SpeechToTextManager(r)
	alpha_rsp := AlphaManager(stt_rsp)
	TextToSpeechManager(alpha_rsp)
}

func SpeechToTextManager(r *http.Request) []byte {
	stt_uri := "http://localhost:3002/stt"

	req, _ := http.NewRequest("POST", stt_uri, r.Body)
	req.Header.Set("Content-Type", "application/json")

	rsp, _ := http.DefaultClient.Do(req)

	if rsp.StatusCode != http.StatusOK {
		panic("something went wrong!")
	}

	defer rsp.Body.Close()

	if rsp.StatusCode == http.StatusOK {
		rsp_body, _ := ioutil.ReadAll(rsp.Body)
		println(string(rsp_body)) // check response is received from speech-to-text microservice
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
		panic("something went wrong!")
	}

	return alpha_rsp.Body
}

func TextToSpeechManager(tts_rsp io.ReadCloser) []byte {
	tts_uri := "http://localhost:3003/tts"

	req, _ := http.NewRequest("POST", tts_uri, tts_rsp)
	req.Header.Set("Content-Type", "application/json")

	rsp, _ := http.DefaultClient.Do(req)

	if rsp.StatusCode != http.StatusOK {
		panic("something went wrong!")
	}

	defer rsp.Body.Close()

	if rsp.StatusCode == http.StatusOK {
		rsp_body, _ := ioutil.ReadAll(rsp.Body)
		println(string(rsp_body)) // check response is received from speech-to-text microservice
		return rsp_body
	}
	return nil
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
