package main

import (
	"encoding/xml"

	"github.com/asticode/go-astisub"

	"github.com/liclac/twibeeo/twiml"
)

func ReadSubFile(filename string) (*twiml.Response, error) {
	subs, err := astisub.OpenFile(filename)
	if err != nil {
		return nil, err
	}

	resp := &twiml.Response{}
	for _, subItem := range subs.Items {
		for _, line := range subItem.Lines {
			for _, item := range line.Items {
				resp.Say = append(resp.Say, twiml.Say{
					Language: "en-US",
					Voice:    "alice",
					Text:     item.Text,
				})

				// Twilio gets sad if we send them more than 64k of TwiML :(
				xmlData, err := xml.Marshal(resp)
				if err != nil {
					return nil, err
				}
				if len(xmlData) > (64*1024)-64 {
					resp.Say = resp.Say[:len(resp.Say)-2]
					return resp, nil
				}
			}
		}
	}
	return resp, nil
}
