package main

import (
	"bufio"
	"context"
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
)

const MFA_TIMEOUT = 30

func main() {

	// fetch url from stdin
	url := get_url()
	color.Cyan(url)

	// login headlessly
	sso_login(url)
	time.Sleep(1 * time.Second)
}

// TODO: time this operation out before `aws sso login`` times out
// returns sso url from stdin.
func get_url() string {
	scanner := bufio.NewScanner(os.Stdin)
	url := ""
	for url == "" {
		scanner.Scan()
		t := scanner.Text()
		// NOTE: adapted from bash: egrep "^https://.*\user_code=([A-Z]{4}-?){2}"
		r, _ := regexp.Compile("^https.*user_code=([A-Z]{4}-?){2}")

		if r.MatchString(t) {
			url = t
		}
	}

	return url
}

// TODO: handle other MFA
// NOTE:
// There are two paths but we always go with the you're not authenticated one :(, we can make it faster somehow
// login with U2f MFA
func sso_login(url string) {

	browser := rod.New().
		MustConnect().
		Trace(false)

	defer browser.MustClose()

	err := rod.Try(func() {

		page := browser.MustPage(url)

		// authorize
		page.MustElementR("button", "Next").MustWaitEnabled().MustPress()
		log.Println(page.MustInfo().Title)

		// sign-in
		signin(*page)

		// TODO: detect if u2f or requires user to type authkey
		// mfa required step
		ssomfa(*page)

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
				//page.MustWaitLoad()
				time.Sleep(1 * time.Second)
			}
		}
	})

	if errors.Is(err, context.DeadlineExceeded) {
		log.Panic("Timeout")
	} else if err != nil {
		log.Panic(err)
	}
}

// executes aws sso signin step
func signin(page rod.Page) {

	usr, _ := user.Current()
	f, _ := netrc.ParseFile(filepath.Join(usr.HomeDir, ".netrc"))

	username := f.FindMachine("headless-sso", "").Login
	page.MustElement("#awsui-input-0").MustInput(username).MustPress(input.Enter)

	passphrase := f.FindMachine("headless-sso", "").Password
	page.MustElement("#awsui-input-1").MustInput(passphrase).MustPress(input.Enter)
	log.Println(page.MustInfo().Title)
}

// TODO: allow user to enter MFA Code
// We only support u2f/hardware MFA right now
// And if you park here, stuff just fails eventually...
func ssomfa(page rod.Page) {

	// we used a page.Race here before but we can't read from stdin after the initial pipe
	log.Println("Touch U2f...")
}
