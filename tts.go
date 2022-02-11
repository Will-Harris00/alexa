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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TextToSpeech(text []byte) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", URI, bytes.NewBuffer(text))

	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
	req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")

	rsp, _ := client.Do(req)

	defer rsp.Body.Close()

	// the request was successful
	if rsp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(rsp.Body)
		return body, nil
	} else {
		CheckStatusError(rsp.StatusCode)
		return nil, errors.New("cannot convert text to speech")
	}
}

func CheckStatusError(err_status int) {
	// error handling for each status code
	if err_status == 400 {
		panic("Bad request - A required parameter is missing, empty, or null. " +
			"Or, the value passed to either a required or optional parameter is invalid. " +
			"A common reason is a header that's too long.")
	}
	if err_status == 401 {
		panic("Unauthorized - The request is not authorized. " +
			"Make sure your subscription key or token is valid and in the correct region.")
	}
	if err_status == 429 {
		panic("Too many requests - You have exceeded the quota or rate of requests allowed for your subscription.")
	}
	if err_status == 502 {
		panic("Bad gateway - There's a network or server-side problem. This status might also indicate invalid headers.")
	}
}

func SpeechEncoding(body []byte) (encoded_speech string) {
	encoded_speech = base64.StdEncoding.EncodeToString(body)
	return encoded_speech
}

func ExtractText(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if answer_text, ok := t["text"].(string); ok {
			println(answer_text)
			ssml := CreateSSML(answer_text) // Speech Synthesis Markup Language
			body, _ := TextToSpeech(ssml)
			TTSResponse(w, body)
		}
	}
}

func TTSResponse(w http.ResponseWriter, body []byte) {
	encoded_speech := SpeechEncoding(body)
	u := map[string]interface{}{"speech": encoded_speech}
	w.Header().Set("Content-Type", "application/json") // return microservice response as json
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func TTSHandler() {
	r := mux.NewRouter()
	// document
	r.HandleFunc("/tts", ExtractText).Methods("POST")
	http.ListenAndServe(":3003", r)
	//	3001 / alpha
	//	3002 / stt
	//	3003 / tts
}

func CreateSSML(text string) []byte {
	speak := &speak{
		Version: "1.0",
		Lang:    "en-US",
		Voice: voice{
			Voice: text,
			Lang:  "en-US",
			Name:  "en-US-JennyNeural",
		},
	}

	text_xml, _ := xml.MarshalIndent(speak, "", "    ")

	// println(string(text_xml))
	return text_xml
}

func main() {
	TTSHandler()
}
