/* eslint-disable */

var fs = require('fs');
var debug = require('debug')('scope:test:action:90-nodes-select');

function clickIfVisible(list, index) {
  var el = list[index++];
  el.isDisplayed(function(err, visible) {
    if (err) {
      debug(err);
    } else if (visible) {
      el.click();
    } else {
      if (index < list.length) {
        clickIfVisible(list, index);
      }
    }
  });
}


function selectNode(browser) {
  debug('select node');
  return browser.elementByCssSelector('.nodes-chart-elements .node:nth-child(1) > g', function(err, el) {
    return el.click();
  });
}


function deselectNode(browser) {
  debug('deselect node');
  return browser.elementByCssSelector('.fa-times', function(err, el) {
    return el.click();
  });
}


module.exports = function(cfg) {

  var startUrl = 'http://' + cfg.host + '/';
  // cfg - The configuration object. args, from the example above.
  return function(browser) {
    // browser is created using wd.promiseRemote()
    // More info about wd at https://github.com/admc/wd
    return browser.get('http://' + cfg.host + '/')
      .then(function() {
        debug('starting run ' + cfg.run);
        return browser.sleep(2000);
      })
      .then(function() {
        return browser.execute("localStorage.debugToolbar = 1;");
      })
      .then(function() {
        return browser.sleep(5000);
      })
      .then(function() {
        return browser.elementByCssSelector('.debug-panel button:nth-child(5)');
        // return browser.elementByCssSelector('.debug-panel div:nth-child(2) button:nth-child(9)');
      })
      .then(function(el) {
        debug('debug-panel found');
        return el.click();
      })
      .then(function() {
        return browser.sleep(2000);
      })
      .then(function() {
        return selectNode(browser);
      })
      .then(function() {
        return browser.sleep(5000);
      })
      .then(function() {
        return deselectNode(browser);
      })
      .then(function() {
        return browser.sleep(2000);
      })
      .then(function() {
        return selectNode(browser);
      })
      .then(function() {
        return browser.sleep(5000);
      })
      .then(function() {
        return deselectNode(browser);
      })
      .then(function() {
        return browser.sleep(2000);
      })
      .then(function() {
        return selectNode(browser);
      })
      .then(function() {
        return browser.sleep(5000);
      })
      .then(function() {
        return deselectNode(browser);
      })

      .then(function() {
        return browser.sleep(2000, function() {
          debug('scenario done');
        });
      })
      .fail(function(err) {
        debug('exception. taking screenshot', err);
        browser.takeScreenshot(function(err, data) {
          if (err) {
            debug(err);
          } else {
            var base64Data = data.replace(/^data:image\/png;base64,/,"");
            fs.writeFile('90-nodes-select-' + cfg.run + '.png', base64Data, 'base64', function(err) {
              if(err) debug(err);
            });
          }
        });
      });
  }
};
