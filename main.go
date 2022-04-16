package main

import (
	"bufio"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fatih/color"
	"github.com/git-lfs/go-netrc/netrc"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// Time before MFA step times out
const MFA_TIMEOUT = 30

func main() {

	// fetch url from stdin
	url := getURL()
	color.Cyan(url)

	// login headlessly
	ssoLogin(url)
	time.Sleep(1 * time.Second)
}

// returns sso url from stdin.
func getURL() string {
	scanner := bufio.NewScanner(os.Stdin)
	url := ""
	for url == "" {
		scanner.Scan()
		t := scanner.Text()
		r, _ := regexp.Compile("^https.*user_code=([A-Z]{4}-?){2}")

		if r.MatchString(t) {
			url = t
		}
	}

	return url
}

// login with U2f MFA
func ssoLogin(url string) {

	browser := rod.New().
		MustConnect().
		Trace(false)

	loadCookies(*browser)

	defer browser.MustClose()

	err := rod.Try(func() {

		page := browser.MustPage(url)

		// authorize
		page.MustElementR("button", "Next").MustWaitEnabled().MustPress()
		log.Println(page.MustInfo().Title)

		// sign-in
		page.Race().ElementR("button", "Allow").MustHandle(func(e *rod.Element) {
		}).Element("#awsui-input-0").MustHandle(func(e *rod.Element) {
			signIn(*page)

			// mfa required step
			mfa(*page)
		}).MustDo()

		// allow request
		unauthorized := true
		for unauthorized {

			txt := page.Timeout(MFA_TIMEOUT * time.Second).MustElement(".awsui-util-mb-s").MustWaitLoad().MustText()
			if txt == "Request approved" {
				log.Println(txt)
				unauthorized = false
			} else {
				exists, _, _ := page.HasR("button", "Allow")
				if exists {
					page.MustWaitLoad().MustElementR("button", "Allow").MustClick()
				}

				time.Sleep(500 * time.Millisecond)
			}
		}

		saveCookies(*browser)
	})

	if errors.Is(err, context.DeadlineExceeded) {
		log.Panic("Timeout")
	} else if err != nil {
		log.Panic(err)
	}
}

// executes aws sso signin step
func signIn(page rod.Page) {
	usr, _ := user.Current()

	f, _ := netrc.ParseFile(filepath.Join(usr.HomeDir, ".netrc"))
	username := f.FindMachine("headless-sso", "").Login
	passphrase := f.FindMachine("headless-sso", "").Password

	page.MustElement("#awsui-input-0").MustInput(username).MustPress(input.Enter)
	page.MustElement("#awsui-input-1").MustInput(passphrase).MustPress(input.Enter)
	log.Println(page.MustInfo().Title)
}

// TODO: allow user to enter MFA Code
func mfa(page rod.Page) {
	log.Println("Touch U2f...")
}

// load cookies
func loadCookies(browser rod.Browser) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	data, _ := os.ReadFile(dirname + "/.headless-sso")
	sEnc, _ := b64.StdEncoding.DecodeString(string(data))
	var cookie *proto.NetworkCookie
	json.Unmarshal(sEnc, &cookie)

	if cookie != nil {
		browser.MustSetCookies(cookie)
	}
}

// save authn cookie
func saveCookies(browser rod.Browser) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	cookies := (browser.MustGetCookies())

	for _, cookie := range cookies {
		if cookie.Name == "x-amz-sso_authn" {
			data, _ := json.Marshal(cookie)

			sEnc := b64.StdEncoding.EncodeToString([]byte(data))
			err = os.WriteFile(dirname+"/.headless-sso", []byte(sEnc), 0644)

			if err != nil {
				log.Panic(err)
			}
			break
		}
	}
}
