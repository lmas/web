
TODO
================================================================================

Big list of temporary notes, to be removed/documented as work continues...

SECURITY 
--------------------------------------------------------------------------------

Validators
- https://www.ssllabs.com/ssltest/
- https://securityheaders.com/
- https://validator.w3.org/
- https://html5.validator.nu/
- https://search.google.com/test/mobile-friendly?hl=en
- http://watson.addy.com/
- https://gf.dev/expect-ct-test
- https://csp-evaluator.withgoogle.com/
- https://www.cspvalidator.org/
- https://metatags.io/

SECURITY HEADERS
--------------------------------------------------------------------------------

example
        X-Frame-Options "deny"
        X-XSS-Protection "1; mode=block"
        X-Content-Type-Options "nosniff"
        Referrer-Policy "origin-when-cross-origin, strict-origin-when-cross-origin"
        Content-Security-Policy "default-src 'self'; style-src 'self' 'unsafe-inline'; object-src 'none'"
        Access-Control-Allow-Origin "https://{host}"
        Expect-CT "max-age=2592000, enforce, report-uri=https://{host}/reports/ct"
        Strict-Transport-Security "max-age=31536000; includeSubDomains"

- https://www.owasp.org/index.php/OWASP_Secure_Headers_Project
- https://github.com/github/secure_headers
- https://github.com/kr/secureheader

content security policy (90 days)
- https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cross-Origin-Embedder-Policy
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referrer-Policy
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
- https://developer.mozilla.org/en-US/docs/Glossary/CORS

seems to be obsolete by june 2021 (acording to MDN)?
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Expect-CT

experimentals?
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cross-Origin-Opener-Policy
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cross-Origin-Resource-Policy
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cross-Origin-Embedder-Policy

AUTH
--------------------------------------------------------------------------------

- https://latacora.micro.blog/2019/07/24/how-not-to.html
- https://latacora.micro.blog/2018/04/03/cryptographic-right-answers.html
- https://github.com/o1egl/paseto

when running TLS, try using js web crypto (get digest of client's password)?
- https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto

TLS
--------------------------------------------------------------------------------

- https://github.com/ssllabs/research/wiki/SSL-and-TLS-Deployment-Best-Practices

let's encrypt/acme auto cert
- https://pkg.go.dev/golang.org/x/crypto/acme/autocert

here's a good HSTS header (90 days) and HTTP redirector example
- https://ssl-config.mozilla.org/#server=go&version=1.14.4&config=modern&guideline=5.6

investigate OCSP stapling support in go
- https://blog.cloudflare.com/high-reliability-ocsp-stapling/
- https://gist.github.com/sleevi/5efe9ef98961ecfb4da8

enable tls SNI (server name indication, basicly multi host sites over tls)
- https://blog.gopheracademy.com/caddy-a-look-inside/
and
- https://github.com/caddyserver/caddy/blob/e42c6bf0bb00d2e5e966ec7d9923eb21627a6b74/server/server.go#L123

OPTIMIZATIONS
--------------------------------------------------------------------------------

- https://methane.github.io/2015/02/reduce-allocation-in-go-code/

DATABASE
--------------------------------------------------------------------------------

- check https://awesome-go.com/ for a sqlite code generator?

EMBEDDING
--------------------------------------------------------------------------------

- https://tip.golang.org/pkg/embed/

MIDDLEWARES
--------------------------------------------------------------------------------

- https://pkg.go.dev/github.com/go-chi/chi@v1.5.1/middleware
- https://pkg.go.dev/github.com/labstack/echo/v4@v4.1.17/middleware
- https://github.com/guardrailsio/awesome-golang-security
- https://github.com/unrolled/secure
- https://caddyserver.com/docs/modules/

