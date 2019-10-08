const webpack = require('webpack');
const autoprefixer = require('autoprefixer');
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');

/**
 * This is the Webpack configuration file for local development.
 * It contains local-specific configuration which includes:
 *
 * - Hot reloading configuration
 * - The entry points of the application
 * - Which loaders to use on what files to properly transpile the source
 *
 * For more information, see: http://webpack.github.io/docs/configuration.html
 */

module.exports = {
  // Efficiently evaluate modules with source maps
  devtool: 'eval-source-map',

  // Set entry points with hot loading
  entry: {
    app: [
      './app/scripts/main',
      'webpack-hot-middleware/client'
    ],
    'dev-app': [
      './app/scripts/main.dev',
      'webpack-hot-middleware/client'
    ],
    'terminal-app': [
      './app/scripts/terminal-main',
      'webpack-hot-middleware/client'
    ],
    vendors: ['@babel/polyfill', 'classnames', 'dagre', 'filesize', 'immutable',
      'moment', 'page', 'react', 'react-dom', 'react-motion', 'react-redux', 'redux',
      'redux-thunk', 'reqwest', 'xterm', 'webpack-hot-middleware/client'
    ]
  },

  // See https://webpack.js.org/concepts/mode/#mode-development.
  mode: 'development',

  // Used by Webpack Dev Middleware
  output: {
    filename: '[name].js',
    path: path.join(__dirname, 'build'),
    publicPath: '',
  },

  // Necessary plugins for hot load
  plugins: [
    new webpack.HotModuleReplacementPlugin(),
    new webpack.NoEmitOnErrorsPlugin(),
    new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/),
    new HtmlWebpackPlugin({
      chunks: ['vendors', 'terminal-app'],
      filename: 'terminal.html',
      template: 'app/html/index.html',
    }),
    new HtmlWebpackPlugin({
      chunks: ['vendors', 'dev-app'],
      filename: 'dev.html',
      template: 'app/html/index.html',
    }),
    new HtmlWebpackPlugin({
      chunks: ['vendors', 'app'],
      filename: 'index.html',
      template: 'app/html/index.html',
    }),
  ],

  // Transform source code using Babel and React Hot Loader
  module: {
    // Webpack is opionated about how pkgs should be laid out:
    // https://github.com/webpack/webpack/issues/1617
    noParse: [/xterm\/(.*).map$/, /xterm\/dist\/xterm\.js/],

    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules|vendor/,
        loaders: [
          'eslint-loader',
          'stylelint-custom-processor-loader',
        ],
        enforce: 'pre'
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
        test: /\.(ttf|eot|svg)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file-loader'
      },
      {
        test: /\.(ico)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file-loader',
        options: {
          // Handle favicon.ico separately as it needs to preserve its name to be applied correctly.
          // See https://github.com/coryhouse/react-slingshot/issues/128#issuecomment-216363426.
          name: '[name].[ext]',
        },
      },
      {
        test: /\.jsx?$/,
        exclude: /node_modules|vendor/,
        use: [
          { loader: 'babel-loader', options: { cacheDirectory: true } },
        ]
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
      }
    ]
  },

  // Automatically transform files with these extensions
  resolve: {
    extensions: ['.js', '.jsx']
  },
};
