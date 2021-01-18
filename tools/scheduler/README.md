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
  $ gcloud app deploy --version $(date '+%Y%m%dt%H%M%S') --project positive-cocoa-90213
  Services to deploy:

  descriptor:      [/Users/simon/weave/build-tools/scheduler/app.yaml]
  source:          [/Users/simon/weave/build-tools/scheduler]
  target project:  [positive-cocoa-90213]
  target service:  [default]
  target version:  [20200512t154238]
  target url:      [https://positive-cocoa-90213.appspot.com]


  Do you want to continue (Y/n)?

  Beginning deployment of service [default]...
  ╔════════════════════════════════════════════════════════════╗
  ╠═ Uploading 433 files to Google Cloud Storage              ═╣
  ╚════════════════════════════════════════════════════════════╝
  File upload done.
  Updating service [default]...done.
  Setting traffic split for service [default]...done.
  Deployed service [default] to [https://positive-cocoa-90213.appspot.com]

  You can stream logs from the command line by running:
    $ gcloud app logs tail -s default

  To view your application in the web browser run:
    $ gcloud app browse --project=positive-cocoa-90213
  ```

- Go to [console.cloud.google.com](https://console.cloud.google.com) > Weave Integration Tests (`positive-cocoa-90213`) > AppEngine > Versions and ensure traffic is being directed to the newly deployed version.
- Click on Tools > Logs, and ensure the application is behaving well.
