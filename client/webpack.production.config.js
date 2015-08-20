var webpack = require('webpack');
var autoprefixer = require('autoprefixer-core');

/**
 * This is the Webpack configuration file for production.
 */
module.exports = {
  entry: './app/scripts/main',

  output: {
    path: __dirname + '/build/',
    filename: 'app.js'
  },

  module: {
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
      { test: /\.jsx?$/, exclude: /node_modules/, loader: 'babel-loader?stage=0' }
    ]
  },

  postcss: [
    autoprefixer({
      browsers: ['last 2 versions']
    })
  ],

  resolve: {
    extensions: ['', '.js', '.jsx']
  },

  plugins: [new webpack.optimize.UglifyJsPlugin({
    compress: {
      warnings: false
    }
  })]
};
