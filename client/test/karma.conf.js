module.exports = function(config) {
  config.set({
    browsers: [
      'PhantomJS'
    ],
    files: [
      './polyfill.js',
      {
        pattern: 'tests.webpack.js',
        watched: false
      }
    ],
    frameworks: [
      'jasmine'
    ],
    preprocessors: {
      'tests.webpack.js': ['webpack']
    },
    reporters: [
      'dots',
      'coverage'
    ],
    coverageReporter: {
      type: 'text-summary'
    },
    webpack: {
      module: {
        loaders: [
          {
            test: /\.js?$/,
            exclude: /node_modules/,
            loader: 'babel-loader'
          }
        ],
        postLoaders: [
          {
            test: /\.js$/,
            exclude: /(test|node_modules|bower_components)\//,
            loader: 'istanbul-instrumenter'
          }
        ]
      },
      watch: true
    },
    webpackServer: {
      noInfo: true
    }
  });
};
