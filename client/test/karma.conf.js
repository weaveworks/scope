module.exports = function(config) {
  config.set({
    browsers: [
      'PhantomJS'
    ],
    files: [
      '../app/**/__tests__/*.js'
    ],
    frameworks: [
      'jasmine', 'browserify'
    ],
    preprocessors: {
      '../app/**/__tests__/*.js': ['browserify']
    },
    browserify: {
      debug: true,
      transform: ['reactify', 'babelify']
    },
    reporters: [
      'dots'
    ]
  });
};
