# Scope UI

## Getting Started (using local node)

- You need nodejs 4.2.2 and a running `weavescope` container
- Setup: `npm install`
- Develop: `BACKEND_HOST=<dockerhost-ip> npm start` and then open `http://localhost:4042/`

This will start a webpack-dev-server that serves the UI and proxies API requests to the container.

## Getting Started (using node in a container)

- You need a running `weavescope` container
- Develop: `make WEBPACK_SERVER_HOST=<dockerhost-ip> client-start` and then open `http://<dockerhost-ip>:4042/`

This will start a webpack-dev-server that serves the UI from the UI build container and proxies API requests to the weavescope container.

## Test Production Bundles Locally

- Build: `npm run build`, output will be in `build/`
- Serve files from `build/`: `BACKEND_HOST=<dockerhost-ip> npm run start-production` and then open `http://localhost:4042/`

## Coding

This directory has a `.eslintrc`, make sure your editor supports linter hints.
To run a linter, you also run `npm run lint`.

## Logging

To enable logging in the console, activate it via `localStorage` in the dev tools console:

```
localStorage["debug"] = "scope:*"
```

The Scope UI uses [debug](https://www.npmjs.com/package/debug) for logging, e.g.,:

```
const debug = require('debug')('scope:app-store');
debug('Store log message');
```
