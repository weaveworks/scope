var webpack = require('webpack');
var autoprefixer = require('autoprefixer-core');
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
var GLOBALS = {
  __WS_URL__: JSON.stringify('ws://' + BACKEND_HOST)
};

module.exports = {

  // Efficiently evaluate modules with source maps
  devtool: 'eval',

  // Set entry point include necessary files for hot load
  entry: [
    'webpack-dev-server/client?http://localhost:4041',
    'webpack/hot/only-dev-server',
    './app/scripts/main'
  ],

  // This will not actually create a app.js file in ./build. It is used
  // by the dev server for dynamic hot loading.
  output: {
    path: path.join(__dirname, 'build/'),
    filename: 'app.js',
    publicPath: 'http://localhost:4041/build/'
  },

  // Necessary plugins for hot load
  plugins: [
    new webpack.DefinePlugin(GLOBALS),
    new webpack.HotModuleReplacementPlugin(),
    new webpack.NoErrorsPlugin()
  ],

  // Transform source code using Babel and React Hot Loader
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
        exclude: /node_modules/,
        loaders: ['react-hot', 'babel-loader?stage=0']
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
