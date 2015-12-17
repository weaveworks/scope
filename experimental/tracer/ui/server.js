var express = require('express')

var app = express();

app.get('/', function(req, res) {
    res.sendFile(__dirname + '/index.html');
});

app.get('/container', function(req, res) {
    res.sendFile(__dirname + '/container.json');
});

app.get('/traces', function(req, res) {
    res.sendFile(__dirname + '/traces.json');
});

app.use(express.static('./'));

var port = process.env.PORT || 4050;
var server = app.listen(port, function () {
    var host = server.address().address;
    var port = server.address().port;

    console.log('Scope Tracer UI listening at http://%s:%s', host, port);
});
