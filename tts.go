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
	KEY = "19c1cb3c0aa848608fed5a5a8a23d640"
)

type Speak struct {
	XMLName xml.Name `xml:"speak"`
	Version string   `xml:"version,attr"`
	Lang    string   `xml:"xml:lang,attr"`
	Voice   struct {
		Text string `xml:",chardata"`
		Lang string `xml:"xml:lang,attr"`
		Name string `xml:"name,attr"`
	} `xml:"voice"`
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
		ResponseError(rsp.StatusCode)
		return nil, errors.New("cannot convert text to speech")
	}
}

func ResponseError(err_status int) {
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

func Text(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if text, ok := t["text"].(string); ok {
			println(text)
			ssml := CreateSSML(text)
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
	r.HandleFunc("/tts", Text).Methods("POST")
	http.ListenAndServe(":3003", r)
	//	3001 / alpha
	//	3002 / stt
	//	3003 / tts
}

func CreateSSML(text string) []byte {
	speak := &Speak{
		Version: "1.0",
		Lang:    "en-US",
		Voice: struct {
			Text string `xml:",chardata"`
			Lang string `xml:"xml:lang,attr"`
			Name string `xml:"name,attr"`
		}(struct {
			Text string
			Lang string
			Name string
		}{
			Text: text,
			Lang: "en-US",
			Name: "en-US-JennyNeural",
		}),
	}

	text_xml, _ := xml.MarshalIndent(speak, "", "    ")

	// print to file if required for testing
	//filename := "text.xml"
	//file, _ := os.Create(filename)
	//
	//xmlWriter := io.Writer(file)
	//
	//enc := xml.NewEncoder(xmlWriter)
	//enc.Indent("", "    ")
	//if err := enc.Encode(text_xml); err != nil {
	//	fmt.Printf("error: %v\n", err)
	//}

	// println(string(text_xml))
	return text_xml
}

func main() {
	TTSHandler()
}
