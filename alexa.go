package main

import (
	"bytes"
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func CheckErr(w http.ResponseWriter, e error, errResp int) {
	if e != nil {
		println(e.Error())
		// println(errResp)
		w.WriteHeader(errResp) // tells our microservice what error response code to return
	}
}

func ProcessAlexa(w http.ResponseWriter, r *http.Request) {
	sttRespBody := SpeechToTextManager(w, r)
	// return
	alphaRespBody := AlphaManager(w, sttRespBody)
	ttsRespBody := TextToSpeechManager(w, alphaRespBody)
	AlexaResponse(w, ttsRespBody)
}

func SpeechToTextManager(w http.ResponseWriter, r *http.Request) []byte {
	sttUri := "http://localhost:3002/stt"

	sttReq, err := http.NewRequest("POST", sttUri, r.Body)
	CheckErr(w, err, http.StatusBadRequest) // the request was malformed

	sttReq.Header.Set("Content-Type", "application/json")

	sttResp, err := http.DefaultClient.Do(sttReq)    // handle error for failed stt microservice query
	CheckErr(w, err, http.StatusInternalServerError) // stt microservice could not be reached
	// return nil

	if sttResp.StatusCode != http.StatusOK {
		err := errors.New("Something went wrong with the speech-to-text microservice!") // handle error for failed stt query
		CheckErr(w, err, sttResp.StatusCode)                                            // pass the stt error code to the alexa microservice response header
	}

	defer sttResp.Body.Close()

	sttRespBody, err := ioutil.ReadAll(sttResp.Body) // read the body of the response returned from the stt microservice
	CheckErr(w, err, http.StatusBadRequest)
	println(string(sttRespBody)) // check response is received from speech-to-text microservice

	return sttRespBody
}

func AlphaManager(w http.ResponseWriter, sttRespBody []byte) []byte {
	alphaUri := "http://localhost:3001/alpha"

	alphaReq, err := http.NewRequest("POST", alphaUri, bytes.NewReader(sttRespBody))
	CheckErr(w, err, http.StatusBadRequest) // the request was malformed

	alphaReq.Header.Set("Content-Type", "application/json")

	alphaResp, err := http.DefaultClient.Do(alphaReq) // handle error for failed alpha query
	CheckErr(w, err, http.StatusInternalServerError)  // alpha microservice could not be reached

	if alphaResp.StatusCode != http.StatusOK {
		err := errors.New("Something went wrong with the alpha microservice!") // handle error for failed stt query
		CheckErr(w, err, alphaResp.StatusCode)                                 // pass the stt error code to the alexa microservice response header
	}

	defer alphaResp.Body.Close()

	alphaRespBody, err := ioutil.ReadAll(alphaResp.Body) // read the body of the response returned from the stt microservice
	CheckErr(w, err, http.StatusBadRequest)
	println(string(alphaRespBody)) // check response is received from speech-to-text microservice

	return alphaRespBody
}

func TextToSpeechManager(w http.ResponseWriter, alphaRespBody []byte) []byte {
	ttsUri := "http://localhost:3003/tts"

	ttsReq, err := http.NewRequest("POST", ttsUri, bytes.NewReader(alphaRespBody))
	CheckErr(w, err, http.StatusBadRequest) // the request was malformed

	ttsReq.Header.Set("Content-Type", "application/json")

	ttsResp, err := http.DefaultClient.Do(ttsReq)    // handle error for failed tts microservice query
	CheckErr(w, err, http.StatusInternalServerError) // tts microservice could not be reached

	if ttsResp.StatusCode != http.StatusOK {
		err := errors.New("Something went wrong with the text-to-speech microservice!") // handle error for failed stt query
		CheckErr(w, err, ttsResp.StatusCode)                                            // pass the stt error code to the alexa microservice response header
	}

	defer ttsResp.Body.Close()

	ttsRespBody, err := ioutil.ReadAll(ttsResp.Body)
	CheckErr(w, err, http.StatusBadRequest) // the server cannot process the request due to something that is perceived to be a client error
	println(string(ttsRespBody))            // check response is received from speech-to-text microservice

	return ttsRespBody
}

func AlexaResponse(w http.ResponseWriter, ttsRespBody []byte) {
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(http.StatusOK)
	w.Write(ttsRespBody)
	println("Play the answer.wav file to hear the solution to your question!")
}

func AlexaHandler() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/alexa", ProcessAlexa).Methods("POST")
	http.ListenAndServe(":3000", r)
	//	3001 / alpha
	//	3002 / stt
	//	3003 / tts
}

func main() {
	AlexaHandler()
}
