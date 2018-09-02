package twiml

import (
	"encoding/xml"
)

type Response struct {
	Say []Say
}

type Say struct {
	Language string `xml:"language,attr,omitempty"`
	Voice    string `xml:"voice,attr,omitempty"`
	Text     string `xml:",chardata"`
	Prosody  []Prosody
}

type Prosody struct {
	XMLName xml.Name `xml:"prosody"`
	Rate    string   `xml:"rate,attr,omitempty"`
	Pitch   string   `xml:"pitch,attr,omitempty"`
	Text    string   `xml:",chardata"`
}
