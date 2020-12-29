# web
[![PkgGoDev](https://pkg.go.dev/badge/github.com/lmas/web)](https://pkg.go.dev/github.com/lmas/web)
[![Tests](https://github.com/lmas/web/workflows/Tests/badge.svg?branch=master)](https://github.com/lmas/web/actions)

Yet another minimal and personal DIY web framework for golang.

## Goals

- Security

        Safe by default, using documented best practices

- Minimalism

        Cherry-picking of features, avoiding 3rd party dependecies

- Sane Defaults

        Minimal configuration required, good performance without ugly hacks

## Status

Under development.

## License

MIT licensed. See the LICENSE file for details.

## References

### Security

- General web security recommendations (feels up to date, lot's of http headers available)

        https://almanac.httparchive.org/en/2020/security
        https://owasp.org/www-project-secure-headers/

- Go web server recommendations (somewhat out of date?)

        https://blog.cloudflare.com/exposing-go-on-the-internet/
        https://juliensalinas.com/en/security-golang-website/

- TLS best practices (and with a good config generator)

        https://wiki.mozilla.org/Security/Server_Side_TLS
        https://www.ssllabs.com/ssl-pulse/

- Go's TLS defaults (cipher suits etc.)

        https://go.googlesource.com/go/+blame/go1.15.6/src/crypto/tls/common.go

- Checking for issues in the source code

        https://github.com/golang/lint
        https://github.com/securego/gosec
