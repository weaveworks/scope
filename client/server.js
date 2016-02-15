var express = require('express');
var httpProxy = require('http-proxy');
var url = require('url');

var app = express();


/************************************************************
 *
 * Express routes for:
 *   - app.js
 *   - app-terminal.js
 *   - index.html
 *
 *   Proxy requests to:
 *     - /api -> :4040/api
 *
 ************************************************************/


// Serve application file depending on environment
app.get(/(app|contrast-app|terminal-app|vendors).js/, function(req, res) {
  var filename = req.originalUrl;
  if (process.env.NODE_ENV === 'production') {
    res.sendFile(__dirname + '/build' + filename);
  } else {
    res.redirect('//localhost:4041/build' + filename);
  }
});

// Proxy to backend

var BACKEND_HOST = process.env.BACKEND_HOST || 'localhost:4040';

var proxy = httpProxy.createProxy({
  ws: true,
  target: 'http://' + BACKEND_HOST
});

proxy.on('error', function(err) {
  console.error('Proxy error', err);
});

app.all('/api*', proxy.web.bind(proxy));

// Serve index page

app.use(express.static('build'));


/*************************************************************
 *
 * Webpack Dev Server
 *
 * See: http://webpack.github.io/docs/webpack-dev-server.html
 *
 *************************************************************/

if (process.env.NODE_ENV !== 'production') {
  var webpack = require('webpack');
  var WebpackDevServer = require('webpack-dev-server');
  var config = require('./webpack.local.config');

  new WebpackDevServer(webpack(config), {
    publicPath: config.output.publicPath,
    hot: true,
    noInfo: true,
    historyApiFallback: true,
    stats: { colors: true }
  }).listen(4041, 'localhost', function (err, result) {
    if (err) {
      console.log(err);
    }
  });
}


/******************
 *
 * Express server
 *
 *****************/

var port = process.env.PORT || 4042;
var server = app.listen(port, function () {
  var host = server.address().address;
  var port = server.address().port;

  console.log('Scope UI listening at http://%s:%s', host, port);
});

server.on('upgrade', proxy.ws.bind(proxy));
