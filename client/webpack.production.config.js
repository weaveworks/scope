const webpack = require('webpack');
const autoprefixer = require('autoprefixer');
const path = require('path');

const CleanWebpackPlugin = require('clean-webpack-plugin');
const ExtractTextPlugin = require('extract-text-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');

const GLOBALS = {
  'process.env': {NODE_ENV: '"production"'}
};

let OUTPUT_PATH = 'build/';
let PUBLIC_PATH = '';

if (process.env.EXTERNAL) {
  OUTPUT_PATH = 'build-external/';
  // Change this line to point to resources on an external host.
  PUBLIC_PATH = 'https://s3.amazonaws.com/static.weave.works/scope-ui/';
}

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
    // keep only some in here, to make vendors and app bundles roughly same size
    vendors: ['babel-polyfill', 'classnames', 'immutable',
      'react', 'react-dom', 'react-redux', 'redux', 'redux-thunk'
    ]
  },

  output: {
    path: path.join(__dirname, OUTPUT_PATH),
    filename: '[name]-[chunkhash].js',
    publicPath: PUBLIC_PATH
  },

  module: {
    // Webpack is opionated about how pkgs should be laid out:
    // https://github.com/webpack/webpack/issues/1617
    noParse: [/xterm\/dist\/xterm\.js/],
    include: [
      path.resolve(__dirname, 'app/scripts', 'app/styles')
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
        test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'url-loader?limit=10000&minetype=application/font-woff'
      },
      {
        test: /\.(ttf|eot|svg|ico)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file-loader'
      },
      {
        test: /\.ico$/,
        loader: 'file-loader?name=[name].[ext]'
      },
      {
        test: /\.jsx?$/,
        exclude: /node_modules|vendor/,
        loader: 'babel'
      },
      {
        test: /\.scss$/,
        loader: ExtractTextPlugin.extract('style-loader', 'css-loader!postcss!sass-loader?minimize')
      }
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
    new CleanWebpackPlugin(['build']),
    new webpack.DefinePlugin(GLOBALS),
    new webpack.optimize.CommonsChunkPlugin('vendors', 'vendors-[chunkhash].js'),
    new webpack.optimize.OccurenceOrderPlugin(true),
    new webpack.IgnorePlugin(/^\.\/locale$/, [/moment$/]),
    new webpack.optimize.UglifyJsPlugin({
      sourceMap: false,
      compress: {
        warnings: false
      }
    }),
    new ExtractTextPlugin('style-[name]-[chunkhash].css'),
    new HtmlWebpackPlugin({
      hash: true,
      chunks: ['vendors', 'contrast-app'],
      template: 'app/html/index.html',
      filename: 'contrast.html'
    }),
    new HtmlWebpackPlugin({
      hash: true,
      chunks: ['vendors', 'terminal-app'],
      template: 'app/html/index.html',
      filename: 'terminal.html'
    }),
    new HtmlWebpackPlugin({
      hash: true,
      chunks: ['vendors', 'app'],
      template: 'app/html/index.html',
      filename: 'index.html'
    })
  ],
  sassLoader: {
    includePaths: [
      path.resolve(__dirname, './node_modules/xterm'),
      path.resolve(__dirname, './node_modules/font-awesome')
    ]
  }
};
