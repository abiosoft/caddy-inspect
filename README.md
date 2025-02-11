# Caddy Inspect

## Inspect HTTP requests in Caddy

![Demonstration](inspect.gif)

## Features

- Insert a breakpoint anywhere in the Caddyfile with the `inspect` keyword.
- Inspecting HTTP and Caddy contexts
- Intercepting and terminating HTTP requests and responses
- Works in the browser, zero setup required.

## Installation

```
xcaddy build --with github.com/abiosoft/caddy-inspect
```

## Getting Started

### Specify breakpoints in the Caddyfile

Place the `inspect` keyword anywhere an [HTTP directive](https://caddyserver.com/docs/caddyfile/directives#caddyfile-directives) is supported in the Caddyfile. The keyword can be specified multiple times.

```caddy
:8080

route /api {
    rewrite /api/* /api/v1{uri}
    inspect
    ...
}
```

### Start Caddy

Caddy can be started with the `--watch` flag to autoreload the Caddyfile on each modification.

```sh
caddy run --watch
```

### Access the Inspect console

Open http://localhost:2020 in the browser to access Caddy Inspect.

> [!NOTE]
> Another port would be auto-assigned if `2020` is not available.
> The URL can be confirmed in the Caddy logs.

### Enjoy

Any HTTP request(s) made to a route containing the `inspect` keyword would pause the request and activate the console.

## Caveats

> [!CAUTION]
> This plugin is intended for development purposes only, sensitive information may be exposed if used in a production environment.

- The plugin is tailored towards Caddyfile config. However, JSON config can be used with limited experience.
- Due to the in-built [order of directives](https://caddyserver.com/docs/caddyfile/directives#directive-order) in Caddy, `inspect` is more predictable in a `route` block. Otherwise, it is ordered after the `encode` directive.
- The information displayed are read-only and cannot be modified.
- HTTP request and response bodies cannot be inspected. It is a deliberate limitation until there is a strong argument in favour.
- The plugin stemmed from a personal use-case. Feedbacks would be appreciated to accommodate more use-cases.

## License

MIT
