# Admin security configuration

Run the site over HTTPS. Authentication cookies are host-only, HttpOnly,
Secure, SameSite=Lax, scoped to `/`, and expire after 20 minutes (the same
lifetime as the in-memory session). `DOMAIN_NAME` no longer controls the cookie
domain. Sessions and login counters are per process and are lost on restart.

For local HTTP development **only**, set `KEIJI_INSECURE_COOKIE=true`. This
removes Secure from the cookie and changes the expected browser origin to HTTP;
never use it on an Internet-facing deployment. TLS termination at a reverse proxy
requires no special cookie configuration: cookies are Secure by default, even if
the proxy connects to Keiji over HTTP. The proxy must preserve the public Host
header and reject unrecognized hosts, prevent direct access to the backend, and
serve only HTTPS. Forwarded host and proto headers are not trusted for CSRF.

Unsafe `/admin` requests and POST `/login` require an `Origin` matching the
public scheme and Host (including any non-default port); if Origin is absent,
Referer is used. Missing, `null`, cross-origin, or malformed values are rejected.
This supports browser forms, multipart uploads, and HTMX without template changes;
a restrictive Referrer-Policy that suppresses both headers will break requests.
Non-browser clients such as keiji-ctl must send an appropriate Origin header on
mutations. Authentication remains required independently. Same-origin XSS is not
prevented by this check.

Login is limited to 10 attempts per IP per 10-minute fixed window, including
successful and malformed attempts; excess requests return 429. By default the
TCP peer is used and forwarding headers are ignored. If a reverse proxy is used,
set `KEIJI_TRUSTED_PROXIES` to a comma-separated list of its IPs or CIDRs, e.g.
`10.0.0.1,192.168.1.0/24`. Only configure addresses under your control: the
limiter walks X-Forwarded-For from the trusted peer to the first untrusted hop.
The proxy must correctly append or overwrite X-Forwarded-For and must not allow
clients to connect directly using a trusted source address. Unconfigured proxies
share a single limit. Invalid entries are ignored. The limiter is in-memory,
per instance, and bounded to 10,000 active IPs; at capacity, new IPs receive
429 until entries expire. For distributed deployments, apply a shared edge rate
limit as well.
