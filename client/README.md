# Scope UI

## Getting Started (using local node)

- You need at least Node.js 6.9.0 and a running `weavescope` container
- Get Yarn: `npm install -g yarn`
- Setup: `yarn install`
- Develop: `BACKEND_HOST=<dockerhost-ip> yarn start` and then open `http://localhost:4042/`

This will start a webpack-dev-server that serves the UI and proxies API requests to the container.

## Getting Started (using node in a container)

- You need a running `weavescope` container
- Develop: `make WEBPACK_SERVER_HOST=<dockerhost-ip> client-start` and then open `http://<dockerhost-ip>:4042/`

This will start a webpack-dev-server that serves the UI from the UI build container and proxies API requests to the weavescope container.

## Test Production Bundles Locally

- Build: `yarn run build`, output will be in `build/`
- Serve files from `build/`: `BACKEND_HOST=<dockerhost-ip> yarn run start-production` and then open `http://localhost:4042/`

## Coding

This directory has a `.eslintrc`, make sure your editor supports linter hints.
To run a linter, you also run `yarn run lint`.

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

## Gotchas

Got a blank screen when loading `http://localhost:4042`?

Make sure you are accessing the right machine:
If you're running `yarn start` on a virtual machine with IP 10.0.0.8, you need to point your browser to `http://10.0.0.8:4042`.
Also, you may need to manually configure the virtual machine to expose ports 4041 (webpack-dev-server) and 4042 (express proxy).
