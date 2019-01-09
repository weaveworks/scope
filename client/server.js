/* eslint no-console: 0 */
const express = require('express');
const http = require('http');
const httpProxy = require('http-proxy');
const HttpProxyRules = require('http-proxy-rules');

const app = express();

const BACKEND_HOST = process.env.BACKEND_HOST || 'localhost';

/*
*
* Proxy requests to:
*   - /api -> :4040/api
*
*/

const backendProxy = httpProxy.createProxy({
  target: `http://${BACKEND_HOST}:4040`,
  ws: true,
});
backendProxy.on('error', err => console.error('Proxy error', err));
app.all('/api*', backendProxy.web.bind(backendProxy));

/*
*
* Production env serves precompiled content from build/
*
*/

if (process.env.NODE_ENV === 'production') {
  app.use(express.static('build'));
}

/*
*
* Webpack Dev Middleware with Hot Reload
*
* See: https://github.com/webpack/webpack-dev-middleware;
*      https://github.com/glenjamin/webpack-hot-middleware
*
*/

if (process.env.NODE_ENV !== 'production') {
  const webpack = require('webpack');
  const webpackMiddleware = require('webpack-dev-middleware');
  const webpackHotMiddleware = require('webpack-hot-middleware');
  const config = require('./webpack.local.config');
  const compiler = webpack(config);

  app.use(webpackMiddleware(compiler, {
    noInfo: true,
    publicPath: config.output.publicPath, // required
    stats: 'errors-only',
  }));

  app.use(webpackHotMiddleware(compiler));
}


/*
*
* Express server
*
*/

const port = process.env.PORT || 4042;
const server = app.listen(port, 'localhost', () => {
  const host = server.address().address;

  console.log('Scope UI listening at http://%s:%s', host, port);
});

server.on('upgrade', backendProxy.ws.bind(backendProxy));


/*
*
* Path proxy server
*
*/

const proxyRules = new HttpProxyRules({
  rules: {
    '/scoped/': `http://localhost:${port}`,
  }
});

const pathProxy = httpProxy.createProxy({ws: true});
pathProxy.on('error', err => console.error('path proxy error', err));
const pathProxyPort = port + 1;
const proxyPathServer = http.createServer((req, res) => {
  const target = proxyRules.match(req);
  if (!target) {
    res.writeHead(500, {'Content-Type': 'text/plain'});
    return res.end('No rules matched! Check out /scoped/');
  }
  return pathProxy.web(req, res, {target});
}).listen(pathProxyPort, 'localhost', () => {
  const pathProxyHost = proxyPathServer.address().address;
  console.log(
    'Scope Proxy Path UI listening at http://%s:%s/scoped/',
    pathProxyHost, pathProxyPort
  );
});

proxyPathServer.on('upgrade', (req, socket, head) => {
  const target = proxyRules.match(req);
  if (target) {
    pathProxy.ws(req, socket, head, {target});
  }
});
