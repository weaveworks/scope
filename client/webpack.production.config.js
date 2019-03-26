const webpack = require('webpack');
const autoprefixer = require('autoprefixer');
const path = require('path');

const CleanWebpackPlugin = require('clean-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const { themeVarsAsScss } = require('weaveworks-ui-components/lib/theme');

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
    'terminal-app': './app/scripts/terminal-main',
    // keep only some in here, to make vendors and app bundles roughly same size
    vendors: ['@babel/polyfill', 'classnames', 'immutable',
      'react', 'react-dom', 'react-redux', 'redux', 'redux-thunk'
    ]
  },


  // See https://webpack.js.org/concepts/mode/#mode-production.
  mode: 'production',

  output: {
    filename: '[name]-[chunkhash].js',
    path: path.join(__dirname, OUTPUT_PATH),
    publicPath: PUBLIC_PATH,
  },

  plugins: [
    new CleanWebpackPlugin([OUTPUT_PATH]),
    new webpack.DefinePlugin(GLOBALS),
    new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/),
    new webpack.IgnorePlugin(/.*\.map$/, /xterm\/lib\/addons/),
    new HtmlWebpackPlugin({
      chunks: ['vendors', 'terminal-app'],
      filename: 'terminal.html',
      hash: true,
      template: 'app/html/index.html',
    }),
    new HtmlWebpackPlugin({
      chunks: ['vendors', 'app'],
      filename: 'index.html',
      hash: true,
      template: 'app/html/index.html',
    }),
  ],

  module: {
    // Webpack is opionated about how pkgs should be laid out:
    // https://github.com/webpack/webpack/issues/1617
    noParse: [/xterm\/dist\/xterm\.js/],

    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules|vendor/,
        loader: 'eslint-loader',
        enforce: 'pre',
        options: {
          failOnError: true
        }
      },
      {
        test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'url-loader',
        options: {
          limit: 10000,
          minetype: 'application/font-woff',
        }
      },
      {
        test: /\.(ttf|eot|svg|ico)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file-loader'
      },
      {
        test: /\.ico$/,
        loader: 'file-loader',
        options: {
          name: '[name].[ext]'
        }
      },
      {
        test: /\.jsx?$/,
        exclude: /node_modules|vendor/,
        loader: 'babel-loader'
      },
      {
        test: /\.css$/,
        use: [
          { loader: 'style-loader' },
          { loader: 'css-loader' },
          {
            loader: 'postcss-loader',
            options: {
              plugins: [
                autoprefixer({
                  browsers: ['last 2 versions']
                })
              ]
            }
          },
        ],
      },
      {
        test: /\.scss$/,
        use: [
          { loader: 'style-loader' },
          { loader: 'css-loader' },
          {
            loader: 'sass-loader',
            options: {
              data: themeVarsAsScss(),
              includePaths: [
                path.resolve(__dirname, './node_modules/xterm'),
                path.resolve(__dirname, './node_modules/rc-slider'),
              ]
            }
          },
        ],
      }
    ]
  },

  resolve: {
    extensions: ['.js', '.jsx']
  },
};
