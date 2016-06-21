# Managing Vendored Dependencies with `gvt`

These operations result in uncommitted changes to your
branch; you will need to commit them as normal. Execute them in the
root of your checkout.

For these changes to take effect, you'll have to `make clean` before running
`make`.

## Installing gvt

    $ go get -u github.com/FiloSottile/gvt

## Adding a Dependency

    ~/service$ gvt fetch example.com/organisation/module

## Updating a Dependency

    ~/service$ gvt update example.com/organisation/module

## Removing a Dependency

    ~/service$ gvt delete example.com/organisation/module
