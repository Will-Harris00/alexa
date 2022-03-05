package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".tts.speech.microsoft.com/" +
		"cognitiveservices/v1"
	KEY = ""
)

type speak struct {
	Version string `xml:"version,attr"`
	Lang    string `xml:"xml:lang,attr"`
	Voice   voice  `xml:"voice"`
}

type voice struct {
	Voice string `xml:",chardata"`
	Lang  string `xml:"xml:lang,attr"`
	Name  string `xml:"name,attr"`
}

func ProcessTTS(w http.ResponseWriter, r *http.Request) {
	answerText, err, errCode := ExtractText(r)
	if err != nil {
		TTSErrResponse(w, err, errCode) // return an error response from the microservice
	}

	textSSML, err, errCode := CreateSSML(answerText)
	if err != nil {
		TTSErrResponse(w, err, errCode) // return an error response from the microservice
	}

	answerSpeech, err, errCode := TextToSpeech(textSSML)
	if err != nil {
		TTSErrResponse(w, err, errCode) // return an error response from the microservice
	} else {
		TTSResponse(w, answerSpeech) // success
	}
}

func ExtractText(r *http.Request) (string, error, int) {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return "", err, http.StatusBadRequest // could not decode json query due to perceived client error
	}

	answerText, ok := t["text"].(string)

	if !ok { // text field is not present
		err = errors.New("Object contains no field 'text'") // handle error for incorrect json object
		return "", err, http.StatusBadRequest
	}

	println(answerText)

	return answerText, nil, 0
}

func TextToSpeech(textSSML []byte) (string, error, int) {
	client := &http.Client{}
	ttsReq, err := http.NewRequest("POST", URI, bytes.NewBuffer(textSSML))
	if err != nil {
		return "", err, http.StatusBadRequest // the request was malformed
	}

	ttsReq.Header.Set("Content-Type", "application/ssml+xml")
	ttsReq.Header.Set("Ocp-Apim-Subscription-Key", KEY)
	ttsReq.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")

	ttsResp, err := client.Do(ttsReq)
	if err != nil {
		return "", err, http.StatusNotFound // microsoft text-to-speech could not be reached
	}

	// the request was not successful
	if ttsResp.StatusCode != http.StatusOK {
		err = CheckTTSStatusErr(ttsResp.StatusCode) // long text error message
		if err != nil {
			return "", err, ttsResp.StatusCode // pass the microsoft tts error code to our own microservice response header
		}
	}

	defer ttsResp.Body.Close() // defer ensures the response body is closed even in case of runtime error during parsing of response

	ttsRespBody, err := ioutil.ReadAll(ttsResp.Body)
	if err != nil {
		return "", err, http.StatusInternalServerError // could not read the body of the response, perceived to be client error
	}

	answerSpeech := base64.StdEncoding.EncodeToString(ttsRespBody) // converts the string to base64 encoded wav

	return answerSpeech, nil, 0
}

func CreateSSML(answerText string) ([]byte, error, int) {
	speak := &speak{
		Version: "1.0",
		Lang:    "en-US",
		Voice: voice{
			Voice: answerText,
			Lang:  "en-US",
			Name:  "en-US-JennyNeural",
		},
	}

	textSSML, err := xml.MarshalIndent(speak, "", "    ") // Speech Synthesis Markup Language
	if err != nil {
		return nil, err, http.StatusBadRequest // could not generate the xml request file due to perceived client error
	}

	// println(string(textSSML))

	return textSSML, nil, 0
}

func CheckTTSStatusErr(errStatus int) error {
	// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-text-to-speech
	// error handling for each status code
	// println(errStatus)
	switch errStatus {
	case http.StatusBadRequest: // 400
		return errors.New("Bad request - A required parameter is missing, empty, or null. " +
			"Or, the value passed to either a required or optional parameter is invalid. " +
			"A common reason is a header that's too long.")
	case http.StatusUnauthorized: // 401
		return errors.New("Unauthorized - The request is not authorized. " +
			"Make sure your subscription key or token is valid and in the correct region.")
	case http.StatusTooManyRequests: // 429
		return errors.New("Too many requests - You have exceeded the quota or rate of requests allowed for your subscription.")
	case http.StatusBadGateway: // 502
		return errors.New("Bad gateway - There's a network or server-side problem. This status might also indicate invalid headers.")
	}
	return errors.New("Microsoft text-to-speech could not determine the specific error - Refer to error status code!")
}

func TTSResponse(w http.ResponseWriter, answer_speech string) {
	w.WriteHeader(http.StatusOK)
	u := map[string]interface{}{"speech": answer_speech}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	json.NewEncoder(w).Encode(u)
}

func TTSErrResponse(w http.ResponseWriter, err error, errCode int) {
	w.WriteHeader(errCode)
	w.Write([]byte(err.Error()))
	w.Header().Set("Content-Type", "text") // return error message as text
	println(errCode)
	println(err.Error()) // display the error message on the console
}

func TTSHandler() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/tts", ProcessTTS).Methods("POST")
	http.ListenAndServe(":3003", r)
	//	3001 / alpha
	//	3002 / stt
	//	3003 / tts
}

func main() {
	TTSHandler()
}
