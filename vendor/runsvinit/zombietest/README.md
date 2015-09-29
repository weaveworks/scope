# zombietest

This directory contains an integration test to prove runsvinit is actually
reaping zombies. `make` builds and executes the test as follows:

1. We produce a linux/amd64 runsvinit binary by setting GOOS/GOARCH and
   invoking the Go compiler. Requires Go 1.5, or Go 1.4 built with the
   appropriate cross-compile options.

2. The build/zombie.c program spawns five zombies and exits. We compile it for
   linux/amd64 via a zombietest-build container. We do this so `make` works
   from a Mac. This requires a working Docker installation.

3. Once we have linux/amd64 runsvinit and zombie binaries, we produce a
   zombietest container via the Dockerfile. That container contains a single
   runit service, /etc/service/zombie, which supervises the zombie binary. We
   provide no default ENTRYPOINT, so we can supply it at runtime.

4. Once the zombietest container is built, we invoke the test.bash script.
   That launches a version of the container with runsvinit set to NOT reap
   zombies, and after 1 second, verifies that zombies exist. Then, it launches
   a version of the container with runsvinit set to reap zombies, and after 1
   second, verifies that no zombies exist.

