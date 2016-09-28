var browserPerf = require('browser-perf');
var options = {
    selenium: 'http://local.docker:4444/wd/hub',
    actions: [require('./custom-action.js')()]
}
browserPerf('http://local.docker:4040/dev.html', function(err, res){
    console.error(err);
    console.log(res);
}, options);
