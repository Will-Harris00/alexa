package main

import (
	"bytes"
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func ProcessAlexa(w http.ResponseWriter, r *http.Request) {
	sttRespBody, err, errCode := SpeechToTextManager(r)
	if err != nil {
		AlexaErrResponse(w, err, errCode) // return an error response from the microservice
	}

	alphaRespBody, err, errCode := AlphaManager(sttRespBody)
	if err != nil {
		AlexaErrResponse(w, err, errCode) // return an error response from the microservice
	}

	ttsRespBody, err, errCode := TextToSpeechManager(alphaRespBody)
	if err != nil {
		AlexaErrResponse(w, err, errCode) // return an error response from the microservice
	} else {
		AlexaResponse(w, ttsRespBody) // success
	}
}

func SpeechToTextManager(r *http.Request) ([]byte, error, int) {
	sttUri := "http://localhost:3002/stt"

	sttReq, err := http.NewRequest("POST", sttUri, r.Body)
	if err != nil {
		return nil, err, http.StatusBadRequest // the request was malformed
	}

	sttReq.Header.Set("Content-Type", "application/json")

	sttResp, err := http.DefaultClient.Do(sttReq) // handle error for failed stt microservice query
	if err != nil {
		return nil, err, http.StatusNotFound // stt microservice could not be reached
	}

	if sttResp.StatusCode != http.StatusOK {
		err = errors.New("The speech-to-text microservice failed to handle the query!") // handle error for failed stt query
		if err != nil {
			return nil, err, sttResp.StatusCode // pass the stt error code to the alexa microservice response header
		}
	}

	defer sttResp.Body.Close()

	sttRespBody, err := ioutil.ReadAll(sttResp.Body) // read the body of the response returned from the stt microservice
	if err != nil {
		return nil, err, http.StatusInternalServerError // could not read the body of the response, perceived to be client error
	}

	println(string(sttRespBody)) // check response is received from speech-to-text microservice

	return sttRespBody, nil, 0
}

func AlphaManager(sttRespBody []byte) ([]byte, error, int) {
	alphaUri := "http://localhost:3001/alpha"

	alphaReq, err := http.NewRequest("POST", alphaUri, bytes.NewReader(sttRespBody))
	if err != nil {
		return nil, err, http.StatusBadRequest // the request was malformed
	}

	alphaReq.Header.Set("Content-Type", "application/json")

	alphaResp, err := http.DefaultClient.Do(alphaReq) // handle error for failed alpha query
	if err != nil {
		return nil, err, http.StatusNotFound // alpha microservice could not be reached
	}

	if alphaResp.StatusCode != http.StatusOK {
		err = errors.New("The alpha microservice failed to handle the query!") // handle error for failed alpha query
		if err != nil {
			return nil, err, alphaResp.StatusCode // pass the alpha error code to the alexa microservice response header
		}
	}

	defer alphaResp.Body.Close()

	alphaRespBody, err := ioutil.ReadAll(alphaResp.Body) // read the body of the response returned from the alpha microservice
	if err != nil {
		return nil, err, http.StatusInternalServerError // could not read the body of the response, perceived to be client error
	}

	println(string(alphaRespBody)) // check response is received from alpha microservice

	return alphaRespBody, nil, 0
}

func TextToSpeechManager(alphaRespBody []byte) ([]byte, error, int) {
	ttsUri := "http://localhost:3003/tts"

	ttsReq, err := http.NewRequest("POST", ttsUri, bytes.NewReader(alphaRespBody))
	if err != nil {
		return nil, err, http.StatusBadRequest // the request was malformed
	}

	ttsReq.Header.Set("Content-Type", "application/json")

	ttsResp, err := http.DefaultClient.Do(ttsReq) // handle error for failed tts microservice query
	if err != nil {
		return nil, err, http.StatusNotFound // tts microservice could not be reached
	}

	if ttsResp.StatusCode != http.StatusOK {
		err = errors.New("The text-to-speech microservice failed to handle the request!") // handle error for failed tts query
		if err != nil {
			return nil, err, ttsResp.StatusCode // pass the tts error code to the alexa microservice response header
		}
	}

	defer ttsResp.Body.Close()

	ttsRespBody, err := ioutil.ReadAll(ttsResp.Body) // read the body of the response returned from the tts microservice
	if err != nil {
		return nil, err, http.StatusInternalServerError // could not read the body of the response, perceived to be client error
	}

	println(string(ttsRespBody)) // check response is received from text-to-speech microservice

	return ttsRespBody, nil, 0
}

func AlexaResponse(w http.ResponseWriter, ttsRespBody []byte) {
	w.WriteHeader(http.StatusOK)
	w.Write(ttsRespBody)
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	println("Play the answer.wav file to hear the solution to your question!")
}

func AlexaErrResponse(w http.ResponseWriter, err error, errCode int) {
	w.WriteHeader(errCode)
	w.Write([]byte(err.Error()))
	w.Header().Set("Content-Type", "text") // return error message as text
	println(errCode)
	println(err.Error()) // display the error message on the console
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
