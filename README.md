## headless-sso
Runs [aws sso login]() headlessly when using the `--no-browser` option.

[![Go Reference](https://pkg.go.dev/badge/github.com/mziyabo/headless-sso.svg)](https://pkg.go.dev/github.com/mziyabo/headless-sso) [![Go Report Card](https://goreportcard.com/badge/github.com/mziyabo/headless-sso)](https://goreportcard.com/report/github.com/mziyabo/headless-sso) 

### Background

We want to avoid leaving the terminal and opening yet another tab and having to click Next next next...

### Install

```bash
go install github.com/mziyabo/headless-sso@latest
```

### Usage:

``` bash
aws sso login  --profile pp --no-browser | headless-sso
```


**Note:** `headless-sso` gets the AWS user credentials from a `.netrc` file with the following format:
 > machine name has to be `headless-sso`

```
machine headless-sso
login <username>
password <something-secret>
```

[![asciicast](https://asciinema.org/a/9n7butmOUwW9oorGmTKdMeRu4.svg)](https://asciinema.org/a/9n7butmOUwW9oorGmTKdMeRu4)

### Limitations:
- Only works with hardware MFA/U2F - Need to read MFA code from stdin.

### Release Notes:
Working but Still WiP, Contributions welcome.

### License:
Apache-2.0