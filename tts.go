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

func CheckErr(w http.ResponseWriter, e error, err_resp int) {
	if e != nil {
		println(e.Error())
		w.WriteHeader(err_resp) // tells our microservice what error response code to return
	}
}

func ExtractText(w http.ResponseWriter, r *http.Request) string {
	t := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	CheckErr(w, err, http.StatusBadRequest) // could not decode json response due to perceived client error

	answer_text, ok := t["text"].(string)

	if !ok {
		err = errors.New("Object contains no field 'text'") // handle error for incorrect json object
		CheckErr(w, err, http.StatusBadRequest)
	}

	println(answer_text)
	return answer_text
}

func TextToSpeech(w http.ResponseWriter, text []byte) string {
	client := &http.Client{}
	req, err := http.NewRequest("POST", URI, bytes.NewBuffer(text))
	CheckErr(w, err, http.StatusBadRequest) // the request was malformed

	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
	req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")

	rsp, err := client.Do(req)
	CheckErr(w, err, http.StatusBadRequest) // the server did not understand the request

	defer rsp.Body.Close() // defer ensures the response body is closed even in case of runtime error during parsing of response

	// the request was not successful
	if rsp.StatusCode != http.StatusOK {
		CheckStatusErr(rsp.StatusCode) // long text error message
		err = errors.New("Cannot convert text to speech!")
		CheckErr(w, err, rsp.StatusCode) // pass the microsoft error code to our own microservice response header
	}

	answer_speech, err := ioutil.ReadAll(rsp.Body)
	CheckErr(w, err, http.StatusBadRequest)                            // the server cannot process the request due to something that is perceived to be a client error
	encoded_speech := base64.StdEncoding.EncodeToString(answer_speech) // converts the string to base64 encoded wav
	return encoded_speech
}

func CreateSSML(w http.ResponseWriter, text string) []byte {
	speak := &speak{
		Version: "1.0",
		Lang:    "en-US",
		Voice: voice{
			Voice: text,
			Lang:  "en-US",
			Name:  "en-US-JennyNeural",
		},
	}

	text_xml, err := xml.MarshalIndent(speak, "", "    ")
	CheckErr(w, err, http.StatusBadRequest) // could not generate the xml request file due to perceived client error

	// println(string(text_xml))
	return text_xml
}

func CheckStatusErr(err_status int) {
	// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-text-to-speech
	// error handling for each status code
	// println(rsp.StatusCode)
	if err_status == http.StatusBadRequest { // 400
		println("Bad request - A required parameter is missing, empty, or null. " +
			"Or, the value passed to either a required or optional parameter is invalid. " +
			"A common reason is a header that's too long.")
	}
	if err_status == http.StatusUnauthorized { // 401
		println("Unauthorized - The request is not authorized. " +
			"Make sure your subscription key or token is valid and in the correct region.")
	}
	if err_status == http.StatusTooManyRequests { // 429
		println("Too many requests - You have exceeded the quota or rate of requests allowed for your subscription.")
	}
	if err_status == http.StatusBadGateway { // 502
		println("Bad gateway - There's a network or server-side problem. This status might also indicate invalid headers.")
	}
}

func TTSResponse(w http.ResponseWriter, answer_speech string) {
	u := map[string]interface{}{"speech": answer_speech}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func ProcessTTS(w http.ResponseWriter, r *http.Request) {
	answer_text := ExtractText(w, r)
	ssml := CreateSSML(w, answer_text) // Speech Synthesis Markup Language
	answer_speech := TextToSpeech(w, ssml)
	TTSResponse(w, answer_speech)
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
