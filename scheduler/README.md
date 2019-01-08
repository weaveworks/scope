# scheduler

## Development

### Dependencies

Download and install:

- the [original App Engine SDK for Python](https://cloud.google.com/appengine/docs/standard/python/download) ([Linux](https://storage.googleapis.com/appengine-sdks/featured/google_appengine_1.9.77.zip), [macOS](https://storage.googleapis.com/appengine-sdks/featured/GoogleAppEngineLauncher-1.9.77.dmg)) -- this should add `appcfg.py` to your `PATH`
- `python` 2.7.x
- `pip`

### Setup

```console
$ pip install -U virtualenv
$ virtualenv --python=$(which python2.7) $TMPDIR/scheduler
$ source $TMPDIR/scheduler/bin/activate
$ pip install -r requirements.txt -t lib
```

## Deployment

- Run:
  ```console
  $ appcfg.py --version $(date '+%Y%m%dt%H%M%S') update .
  XX:XX PM Application: positive-cocoa-90213; version: 1
  XX:XX PM Host: appengine.google.com
  XX:XX PM Starting update of app: positive-cocoa-90213, version: 1
  XX:XX PM Getting current resource limits.
  Your browser has been opened to visit:

      https://accounts.google.com/o/oauth2/auth?scope=...

  If your browser is on a different machine then exit and re-run this
  application with the command-line parameter

    --noauth_local_webserver

  Authentication successful.
  XX:XX PM Scanning files on local disk.
  XX:XX PM Scanned 500 files.
  XX:XX PM Scanned 1000 files.
  XX:XX PM Cloning 1220 application files.
  XX:XX PM Uploading 28 files and blobs.
  XX:XX PM Uploaded 28 files and blobs.
  XX:XX PM Compilation starting.
  XX:XX PM Compilation completed.
  XX:XX PM Starting deployment.
  XX:XX PM Checking if deployment succeeded.
  XX:XX PM Will check again in 1 seconds.
  XX:XX PM Checking if deployment succeeded.
  XX:XX PM Will check again in 2 seconds.
  XX:XX PM Checking if deployment succeeded.
  XX:XX PM Will check again in 4 seconds.
  XX:XX PM Checking if deployment succeeded.
  XX:XX PM Deployment successful.
  XX:XX PM Checking if updated app version is serving.
  XX:XX PM Completed update of app: positive-cocoa-90213, version: 1
  XX:XX PM Uploading cron entries.
  ```

- Go to [console.cloud.google.com](https://console.cloud.google.com) > Weave Integration Tests (`positive-cocoa-90213`) > AppEngine > Versions and ensure traffic is being directed to the newly deployed version.
- Click on Tools > Logs, and ensure the application is behaving well.
