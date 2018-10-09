# Use of vendored code in Weave Scope

Weave Scope is licensed under the [Apache 2.0 license](LICENSE).

Some vendored code is under different licenses though, all of them ship the
entire license text they are under.

- https://github.com/weaveworks/go-checkpoint  
  https://github.com/weaveworks/go-cleanhttp  
  https://github.com/certifi/gocertifi  
  can be found in the ./vendor/ directory, is under MPL-2.0.

- Pulled in by dependencies are  
  https://github.com/hashicorp/go-version (MPL-2.0)//comment added  
  https://github.com/hashicorp/golang-lru (MPL-2.0)

- One file pulled in by a dependency is under CDDL:  
  ./vendor/github.com/howeyc/gopass/terminal_solaris.go

- The docs of a dependency that's pulled in by a dependency
  are under CC-BY 4.0:
  ./vendor/github.com/docker/go-units/

[One file used in tests](COPYING.LGPL-3) is under LGPL-3, that's why we ship
the license text in this repository.
