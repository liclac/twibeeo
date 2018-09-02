package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"

	"github.com/pkg/errors"

	_ "github.com/joho/godotenv/autoload"
)

var (
	fAddr    = pflag.StringP("addr", "a", env("ROBO_ADDR", "0.0.0.0:8080"), "bind address")
	fBaseURL = pflag.StringP("base-url", "B", env("ROBO_BASE_URL", ""), "base url we're serving on")
	fSubs    = pflag.StringP("subs", "s", env("ROBO_SUBS", ""), "subtitle file to read")

	fDial       = pflag.StringP("dial", "d", env("ROBO_DIAL", ""), "dial a phone number")
	fDialFrom   = pflag.StringP("dial-from", "f", env("ROBO_DIAL_FROM", ""), "number to call from")
	fDialDigits = pflag.StringP("dial-digits", "D", env("ROBO_DIAL_DIGITS", ""), "(optional) DTMF")
	fAccountSID = pflag.StringP("account-sid", "S", env("ROBO_ACCOUNT_SID", ""), "twilio account SID")
	fAuthToken  = pflag.StringP("auth-token", "A", env("ROBO_AUTH_TOKEN", ""), "twilio auth token")
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Main() error {
	pflag.Parse()
	if *fBaseURL == "" {
		return errors.Errorf("--base-url is required")
	}

	twimlData := TheEntireBeeMovieTwiML
	if *fSubs != "" {
		d, err := ReadSubFile(*fSubs)
		if err != nil {
			return err
		}
		twimlData = d
	}

	http.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(rw, "hi!")
	}))

	twimlHit := make(chan time.Time, 1)
	http.Handle("/twiml", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		data, err := xml.Marshal(twimlData)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(rw, err)
			return
		}
		rw.Header().Set("Content-Type", "text/xml")
		io.Copy(rw, bytes.NewReader(data))
		twimlHit <- time.Now()
	}))

	lis, err := net.Listen("tcp", *fAddr)
	if err != nil {
		return err
	}
	log.Printf("Listening on: http://%s\n", lis.Addr())
	go func() {
		if err := http.Serve(lis, nil); err != nil {
			panic(err)
		}
	}()

	ctx := context.Background()
	switch {
	case *fDial != "":
		log.Printf("Starting a Twilio call...")
		twilio := NewTwilioClient(*fAccountSID, *fAuthToken)
		twilio.Debug = true
		req := StartCallRequest{
			From:       *fDialFrom,
			To:         *fDial,
			SendDigits: *fDialDigits,
			Url:        *fBaseURL + "/twiml",
			Method:     "GET",
			Record:     true,
		}
		if _, err := twilio.StartCall(context.Background(), req); err != nil {
			return err
		}

		log.Printf("Waiting for victim to pick up...")
		select {
		case <-ctx.Done():
			log.Printf("Context terminated, shutting down...")
		case <-twimlHit:
			log.Printf("-> Phone picked up, TwiML sent!")
		}
		return nil
	default:
		log.Printf("--dial not specified, doing nothing")
		for {
			select {
			case t := <-twimlHit:
				log.Printf("TwiML served: %s", t)
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
