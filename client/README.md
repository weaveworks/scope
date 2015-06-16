# Scope UI

## Getting Started

- Setup: `npm install`
- Build: `gulp build --release`, output will be in `build/`
- Develop: `gulp sync` and then open `http://localhost:4042/`

To see a topology, `../app/app` needs to be running, as well as a probe.

## Coding

This directory has a `.eslintrc`, make sure your editor supports linter hints.
To run a linter, you also run `gulp lint`.

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
