# Managing Vendored Dependencies with `gvt`

These operations result in uncommitted changes to your
branch; you will need to commit them as normal. Execute them in the
root of your checkout.

## Adding a New Dependency

    ~/scope$ gvt fetch example.com/organisation/module vendor/example.com/organisation/module

## Updateing a Specific Dependancy

    ~/scope$ gvt update example.com/organisation/module vendor/example.com/organisation/module

## Remove a Dependency

    ~/scope$ gvt delete example.com/organisation/module vendor/example.com/organisation/module
