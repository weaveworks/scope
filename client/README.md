# Scope UI

## Requirements

- nodejs 4.2.2
- running `weavescope` container

## Getting Started

- Setup: `npm install`
- Build: `npm run build`, output will be in `build/`
- Develop: `BACKEND_HOST=<dockerhost-ip>:4040 npm start` and then open `http://localhost:4042/`

This will start a webpack-dev-server that serves the UI and proxies API requests to the container.

## Coding

This directory has a `.eslintrc`, make sure your editor supports linter hints.
To run a linter, you also run `npm run lint`.

## Logging

The Scope UI uses [debug](https://www.npmjs.com/package/debug) for logging, e.g.,:

```
const debug = require('debug')('scope:app-store');
debug('Store log message');
```

To enable logging in the console, activate it via `localStorage` in the dev tools console:

```
localStorage["debug"] = "scope:*"
```
