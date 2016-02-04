var webpack = require('webpack');
var autoprefixer = require('autoprefixer');
var path = require('path');

/**
 * This is the Webpack configuration file for local development. It contains
 * local-specific configuration such as the React Hot Loader, as well as:
 *
 * - The entry point of the application
 * - Where the output file should be
 * - Which loaders to use on what files to properly transpile the source
 *
 * For more information, see: http://webpack.github.io/docs/configuration.html
 */

 // Inject websocket url to dev backend
var BACKEND_HOST = process.env.BACKEND_HOST || 'localhost:4040';

module.exports = {

  // Efficiently evaluate modules with source maps
  devtool: 'cheap-module-source-map',

  // Set entry point include necessary files for hot load
  entry: {
    'app': [
      './app/scripts/main',
      'webpack-dev-server/client?http://localhost:4041',
      'webpack/hot/only-dev-server'
    ],
    'terminal-app': [
      './app/scripts/terminal-main',
      'webpack-dev-server/client?http://localhost:4041',
      'webpack/hot/only-dev-server'
    ],
    'components-app': [
      './app/scripts/components-main',
      'webpack-dev-server/client?http://localhost:4041',
      'webpack/hot/only-dev-server'
    ],
    vendors: ['classnames', 'd3', 'dagre', 'flux', 'immutable',
      'lodash', 'page', 'react', 'react-dom', 'react-motion']
  },

  // This will not actually create a app.js file in ./build. It is used
  // by the dev server for dynamic hot loading.
  output: {
    path: path.join(__dirname, 'build/'),
    filename: '[name].js',
    publicPath: 'http://localhost:4041/build/'
  },

  // Necessary plugins for hot load
  plugins: [
    new webpack.optimize.CommonsChunkPlugin('vendors', 'vendors.js'),
    new webpack.HotModuleReplacementPlugin(),
    new webpack.NoErrorsPlugin()
  ],

  // Transform source code using Babel and React Hot Loader
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
        test: /\.json$/,
        loader: 'json-loader'
      },
      {
        test: /\.less$/,
        loader: 'style-loader!css-loader!postcss-loader!less-loader'
      },
      {
        test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'url-loader?limit=10000&minetype=application/font-woff'
      },
      {
        test: /\.(ttf|eot|svg)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file-loader'
      },
      {
        test: /\.jsx?$/,
        exclude: /node_modules|vendor/,
        loaders: ['react-hot', 'babel']
      }
    ]
  },

  postcss: [
    autoprefixer({
      browsers: ['last 2 versions']
    })
  ],

  // Automatically transform files with these extensions
  resolve: {
    extensions: ['', '.js', '.jsx']
  }
};
