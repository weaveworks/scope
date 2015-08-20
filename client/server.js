var express = require('express');
var proxy = require('proxy-middleware');
var httpProxy = require('express-http-proxy');
var url = require('url');

var app = express();


/************************************************************
 *
 * Express routes for:
 *   - app.js
 *   - index.html
 *
 *   Proxy requests to:
 *     - /api -> :4040/api
 *
 ************************************************************/

// Serve application file depending on environment
app.get('/app.js', function(req, res) {
  if (process.env.NODE_ENV === 'production') {
    res.sendFile(__dirname + '/build/app.js');
  } else {
    res.redirect('//localhost:4041/build/app.js');
  }
});

// Proxy to backend

// HACK need express-http-proxy, because proxy-middleware does
// not proxy to /api itself
app.use(httpProxy('localhost:4040', {
  filter: function(req) {
    return url.parse(req.url).path === '/api';
  },
  forwardPath: function(req) {
    return url.parse(req.url).path;
  }
}));

app.use('/api', proxy('http://localhost:4040/api/'));

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
