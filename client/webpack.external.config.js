const webpack = require('webpack');
const autoprefixer = require('autoprefixer');
const path = require('path');

const CleanWebpackPlugin = require('clean-webpack-plugin');
const ExtractTextPlugin = require('extract-text-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');

const GLOBALS = {
  'process.env': {NODE_ENV: '"production"'}
};

/**
 * This is the Webpack configuration file for hosting most of the static content
 * (all but index.htm) on an external host (eg. a CDN).
 * You should change output.publicPath to point to your external host.
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
    vendors: ['babel-polyfill', 'classnames', 'd3', 'immutable',
      'lodash', 'react', 'react-dom', 'react-redux',
      'redux', 'redux-thunk']
  },

  output: {
    path: path.join(__dirname, 'build-external/'),
    filename: '[name]-[chunkhash].js',
    // Change this line to point to resources on an external host.
    publicPath: 'https://s3.amazonaws.com/static.weave.works/scope-ui/'
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
        loader: ExtractTextPlugin.extract('style-loader',
          'css-loader?minimize!postcss-loader!less-loader')
      },
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
  ]
};
