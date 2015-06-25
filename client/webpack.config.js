/*
 * Webpack config based on https://github.com/kriasoft/react-starter-kit
 */

'use strict';

var autoprefixer = require('autoprefixer-core');
var _ = require('lodash');
var webpack = require('webpack');
var argv = require('minimist')(process.argv.slice(2));

var DEBUG = !argv.release;
var STYLE_LOADER = 'style-loader';
var CSS_LOADER = DEBUG ? 'css-loader' : 'css-loader?minimize';
var AUTOPREFIXER_LOADER = 'postcss-loader';

//
// Common configuration chunk to be used for both
// client-side (app.js) and server-side (server.js) bundles
// -----------------------------------------------------------------------------

var config = {
  output: {
    path: './build/',
    publicPath: './',
    sourcePrefix: '  '
  },

  bail: !DEBUG, // fail on first error when building release
  cache: DEBUG,
  debug: DEBUG,
  devtool: DEBUG ? '#inline-source-map' : false,

  stats: {
    colors: true,
    reasons: DEBUG
  },

  plugins: [
    new webpack.optimize.OccurenceOrderPlugin()
  ],

  resolve: {
    extensions: ['', '.webpack.js', '.web.js', '.js', '.jsx']
  },

  module: {
    preLoaders: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        loader: 'eslint-loader'
      }
    ],

    loaders: [
      {
        test: /\.css$/,
        loader: STYLE_LOADER + '!' + CSS_LOADER + '!' + AUTOPREFIXER_LOADER
      },
      {
        test: /\.less$/,
        loader: STYLE_LOADER + '!' + CSS_LOADER + '!' + AUTOPREFIXER_LOADER +
                '!less-loader'
      },
      {
        test: /\.gif/,
        loader: 'url-loader?limit=10000&mimetype=image/gif'
      },
      {
        test: /\.jpg/,
        loader: 'url-loader?limit=10000&mimetype=image/jpg'
      },
      {
        test: /\.png/,
        loader: 'url-loader?limit=10000&mimetype=image/png'
      },
      {
        test: /\.svg/,
        loader: 'url-loader?limit=10000&mimetype=image/svg+xml'
      },
      {
        test: /\.jsx?$/,
        exclude: /node_modules/,
        loader: 'babel-loader'
      },
      {
        test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'url-loader?limit=10000&minetype=application/font-woff'
      },
      {
        test: /\.(ttf|eot|svg)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file-loader'
      }
    ]
  },

  postcss: [
    autoprefixer({
      browsers: ['last 2 version']
    })
  ]
};

//
// Configuration for the client-side bundle (app.js)
// -----------------------------------------------------------------------------

var appConfig = _.merge({}, config, {
  entry: './app/scripts/main.js',
  output: {
    filename: 'app.js'
  },
  plugins: config.plugins.concat(DEBUG ? [] : [
      new webpack.optimize.DedupePlugin(),
      new webpack.optimize.UglifyJsPlugin({compress: {warnings: false}}),
      new webpack.optimize.AggressiveMergingPlugin()
    ]
  )
});

module.exports = [appConfig];
