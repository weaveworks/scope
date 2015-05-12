# Scope

## Overview

TODO

## Developing

### Building

To build everything in-place,

```
make build
```

Note that this doesn't build or include the latest version of the user
interface. The UI is decoupled, living in `client` and following a node/gulp
workflow. To build that and include it in the application binary,

```
make client
make static
make build
```

Or, as a shortcut,

```
make dist
```

