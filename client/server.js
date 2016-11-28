var express = require('express');
var http = require('http');
var httpProxy = require('http-proxy');
var HttpProxyRules = require('http-proxy-rules');
var url = require('url');

var app = express();

var BACKEND_HOST = process.env.BACKEND_HOST || 'localhost';
var WEBPACK_SERVER_HOST = process.env.WEBPACK_SERVER_HOST || 'localhost';

/************************************************************
 *
 * Proxy requests to:
 *   - /api -> :4040/api
 *
 ************************************************************/

var backendProxy = httpProxy.createProxy({
  ws: true,
  target: 'http://' + BACKEND_HOST + ':4040'
});
backendProxy.on('error', function(err) {
  console.error('Proxy error', err);
});
app.all('/api*', backendProxy.web.bind(backendProxy));

/************************************************************
 *
 * Production env serves precompiled content from build/
 *
 ************************************************************/

if (process.env.NODE_ENV === 'production') {
  app.use(express.static('build'));
}

/*************************************************************
 *
 * Webpack Dev Middleware with Hot Reload
 *
 * See: https://github.com/webpack/webpack-dev-middleware;
 *      https://github.com/glenjamin/webpack-hot-middleware
 *
 *************************************************************/

if (process.env.NODE_ENV !== 'production') {
  var webpack = require('webpack');
  var webpackMiddleware = require('webpack-dev-middleware');
  var webpackHotMiddleware = require('webpack-hot-middleware');
  var config = require('./webpack.local.config');
  var compiler = webpack(config);

  app.use(webpackMiddleware(compiler, {
    // required
    publicPath: config.output.publicPath,
    // options
    noInfo: true,
    watchOptions: {
      aggregateTimeout: 500,
      poll: true
    },
    stats: 'errors-only',
  }));

  app.use(webpackHotMiddleware(compiler));
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

server.on('upgrade', backendProxy.ws.bind(backendProxy));


/*************************************************************
 *
 * Path proxy server
 *
 *************************************************************/

var proxyRules = new HttpProxyRules({
  rules: {
    '/scoped/': 'http://localhost:' + port
  }
});

var pathProxy = httpProxy.createProxy({ws: true});
pathProxy.on('error', function(err) { console.error('path proxy error', err); });
var pathProxyPort = port + 1;
const proxyPathServer = http.createServer(function(req, res) {
  var target = proxyRules.match(req);
  if (!target) {
    res.writeHead(500, {'Content-Type': 'text/plain'});
    res.end('No rules matched! Check out /scoped/');
    return;
  }
  return pathProxy.web(req, res, {target: target});
}).listen(pathProxyPort, function() {
  var pathProxyHost = proxyPathServer.address().address;
  console.log('Scope Proxy Path UI listening at http://%s:%s/scoped/',
              pathProxyHost, pathProxyPort);
});

proxyPathServer.on('upgrade', function(req, socket, head) {
  var target = proxyRules.match(req);
  if (target) {
    return pathProxy.ws(req, socket, head, {target: target});
  }
});
