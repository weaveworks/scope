var perfjankie = require('perfjankie');

var run = process.env.COMMIT || 'commit#Hash'; // A hash for the commit, displayed in the x-axis in the dashboard
var time = process.env.DATE || new Date().getTime() // Used to sort the data when displaying graph. Can be the time when a commit was made
var scenario = process.env.ACTIONS || '90-nodes-select';
var host = process.env.HOST || 'localhost:4040';
var actions = require('../actions/' + scenario)({host: host, run: run});

perfjankie({
  /* The next set of values identify the test */
  suite: 'Scope',
  name: scenario, // A friendly name for the URL. This is shown as component name in the dashboard
  time: time,
  run: run,
  repeat: 10, // Run the tests 10 times. Default is 1 time

  /* Identifies where the data and the dashboard are saved */
  couch: {
    server: 'http://local.docker:5984',
    database: 'performance'
    // updateSite: !process.env.CI, // If true, updates the couchApp that shows the dashboard. Set to false in when running Continuous integration, run this the first time using command line.
    // onlyUpdateSite: false // No data to upload, just update the site. Recommended to do from dev box as couchDB instance may require special access to create views.
  },

  callback: function(err, res) {
    if (err)
      console.log(err);
    // The callback function, err is falsy if all of the following happen
    // 1. Browsers perf tests ran
    // 2. Data has been saved in couchDB
    // err is not falsy even if update site fails.
  },

  /* OPTIONS PASSED TO BROWSER-PERF  */
  // Properties identifying the test environment */
  browsers: [{ // This can also be a ['chrome', 'firefox'] or 'chrome,firefox'
    browserName: 'chrome',
    chromeOptions: {
      perfLoggingPrefs: {
        'traceCategories': 'toplevel,disabled-by-default-devtools.timeline.frame,blink.console,disabled-by-default-devtools.timeline,benchmark'
      },
      args: ['--enable-gpu-benchmarking', '--enable-thread-composting']
    },
    loggingPrefs: {
      performance: 'ALL'
    }
  }], // See browser perf browser configuration for all options.
  actions: actions,

  selenium: {
    hostname: 'local.docker', // or localhost or hub.browserstack.com
    port: 4444,
  }

});
