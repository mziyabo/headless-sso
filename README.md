## headless-sso
Runs [aws sso login]() headlessly when using the `--no-browser` option.

### Background

We want to avoid leaving the terminal and opening yet another tab and having to click Next next next...

### Installation and Usage:

Usage:

``` bash
aws sso login  --profile pp --no-browser | go run . --
```

We get the AWS credentials from a `.netrc` file with the following format:

> machine name has to be headless-sso

```
machine headless-sso
login <username>
password <something-secret>
```



### Known Issues:
- Only works with hardware MFA/u2f - Need to read MFA code from stdin.
- Occasionally contexts failure - Error handling needs work

### Release Notes:
Working but Still WiP, Contributions welcome.

### License:
Apache-2.0