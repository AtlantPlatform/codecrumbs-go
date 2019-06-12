## codecrumbs-go

An alternative to [codecrumbs.io](https://github.com/Bogdan-Lyashenko/codecrumbs), written in pure Go. The aim of this project is to experiment with the same ideas that the original CC broughts, but using a flexible textual representation engine. The main goal that we seek is to visualise and document our large Go codebases, without tedious maintenance of interactive UI.

### Installation

With Go:

```
$ go get -u github.com/AtlantPlatform/codecrumbs-go/cmd/cc-go
```

Or get a release for your platform from the [releases](https://github.com/AtlantPlatform/codecrumbs-go/releases) section.

### Project Goals

* Be fast and in pure Go;
* Deliver results in predictable fasion;
* Render static documentation in Markdown or using output modules (allowing to yield JSON, UML, etc);
* Rely on external source of code (i.e. redirect user to GitHub or local files).

This project might have an interactive UI in the future, but now it is designed to provide a human readable documentation with code references, using the same syntax of code comments as for the original CodeCrumbs.

### GitHub Rate Limit

This error can occur if you're rendering your Markdown docs too frequently using `cc-go render`. You can switch to standalone renderer (e.g. [pandoc](https://github.com/jgm/pandoc)) or authenticate.

```
github API (403): {
    "message": "API rate limit exceeded for 183.89.199.208. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.)",
    "documentation_url": "https://developer.github.com/v3/#rate-limiting"
}
```

Solve this by registering your own app (https://github.com/settings/applications/new) and providing the keys to `cc-go render`:

```
Usage: cc-go render [OPTIONS] FILE

Renders generated files into some representation (e.g. Markdown -> HTML)

      --client-id       GitHub Client ID for Authorization of requests.
      --client-secret   GitHub Client Secret for Authorization of requests.
```

Example:

```
$ cc-go render --client-id=XXX --client-secret=YYY kek.md
```

### LICENSE

MIT
