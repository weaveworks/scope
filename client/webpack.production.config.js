var webpack = require('webpack');
var autoprefixer = require('autoprefixer');
var path = require('path');

var GLOBALS = {
  'process.env': {NODE_ENV: '"production"'}
};

/**
 * This is the Webpack configuration file for production.
 */
module.exports = {

  // fail on first error when building release
  bail: true,

  cache: {},

  entry: {
    app: './app/scripts/main',
    'contrast-app': './app/scripts/contrast-main',
    'terminal-app': './app/scripts/terminal-main',
    vendors: ['classnames', 'd3', 'dagre', 'flux', 'immutable',
      'lodash', 'page', 'react', 'react-dom', 'react-motion']
  },

  output: {
    path: path.join(__dirname, 'build/'),
    filename: '[name].js'
  },

  module: {
    include: [
      path.resolve(__dirname, 'app/scripts')
    ],
    preLoaders: [
      {
        test: /\.js$/,
        exclude: /node_modules|vendor/,
        loader: 'eslint-loader'
      }
    ],
    loaders: [
      {
        test: /\.less$/,
        loader: 'style-loader!css-loader?minimize!postcss-loader!less-loader'
      },
      {
        test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'url-loader?limit=10000&minetype=application/font-woff'
      },
      {
        test: /\.(ttf|eot|svg)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file-loader'
      },
      { test: /\.jsx?$/, exclude: /node_modules|vendor/, loader: 'babel' }
    ]
  },

  postcss: [
    autoprefixer({
      browsers: ['last 2 versions']
    })
  ],

  eslint: {
    failOnError: true
  },

  resolve: {
    extensions: ['', '.js', '.jsx']
  },

  plugins: [
    new webpack.DefinePlugin(GLOBALS),
    new webpack.optimize.CommonsChunkPlugin('vendors', 'vendors.js'),
    new webpack.optimize.OccurenceOrderPlugin(true),
    new webpack.optimize.UglifyJsPlugin({
      sourceMap: false,
      compress: {
        warnings: false
      }
    })
  ]
};
