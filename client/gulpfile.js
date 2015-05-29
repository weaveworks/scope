/*
 * Gulpfile based on https://github.com/kriasoft/react-starter-kit
 */

'use strict';

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();
var del = require('del');
var path = require('path');
var runSequence = require('run-sequence');
var webpack = require('webpack');
var argv = require('minimist')(process.argv.slice(2));

var watch = false;
var browserSync;

// The default task
gulp.task('default', ['build']);

// Clean output directory
gulp.task('clean', del.bind(
  null, ['.tmp', 'build/*'], {dot: true}
));

// 3rd party libraries
gulp.task('vendor', function() {
  return gulp.src('node_modules/font-awesome/fonts/**')
    .pipe($.filter('**/*.{eot,svg,ttf,woff,woff2}'))
    .pipe(gulp.dest('build/fonts'));
});

// Favicon
gulp.task('favicon', function() {
  return gulp.src(['app/favicon.ico'])
    .pipe(gulp.dest('build'));
});


// Static files
gulp.task('html', function() {
  var release = !!argv.release;

  return gulp.src('app/*.html')
    .pipe($.changed('build'))
    .pipe($.if(release, $.preprocess()))
    .pipe(gulp.dest('build'))
    .pipe($.size({title: 'html'}));
});

// Bundle
gulp.task('bundle', function(cb) {
  var started = false;
  var config = require('./webpack.config.js');
  var bundler = webpack(config);
  var verbose = !!argv.verbose;

  function bundle(err, stats) {
    if (err) {
      throw new $.util.PluginError('webpack', err);
    }

    console.log(stats.toString({
      colors: $.util.colors.supportsColor,
      hash: verbose,
      version: verbose,
      timings: verbose,
      chunks: verbose,
      chunkModules: verbose,
      cached: verbose,
      cachedAssets: verbose
    }));

    if (!started) {
      started = true;
      return cb();
    }
  }

  if (watch) {
    bundler.watch(200, bundle);
  } else {
    bundler.run(bundle);
  }
});

// Build the app from source code
gulp.task('build', ['clean'], function(cb) {
  runSequence(['vendor', 'html', 'favicon', 'bundle'], cb);
});

// Build and start watching for modifications
gulp.task('build:watch', function(cb) {
  watch = true;
  runSequence('build', function() {
    gulp.watch('app/*.html', ['html']);
    cb();
  });
});

// Launch a Node.js/Express server
gulp.task('serve', ['build:watch'], function() {
  $.connect.server({
    root: ['build'],
    port: 4041,
    middleware: function() {
      return [(function() {
        var url = require('url');
        var proxy = require('proxy-middleware');
        var options = url.parse('http://localhost:4040/api');
        options.route = '/api';
        return proxy(options);
      })()];
    },
    livereload: false
  });
});

// Launch BrowserSync development server
gulp.task('sync', ['serve'], function(cb) {
  browserSync = require('browser-sync');

  browserSync({
    logPrefix: 'RSK',
    // Stop the browser from automatically opening
    open: false,
    notify: false,
    // Run as an https by setting 'https: true'
    // Note: this uses an unsigned certificate which on first access
    //       will present a certificate warning in the browser.
    https: false,
    // Informs browser-sync to proxy our Express app which would run
    // at the following location
    proxy: 'localhost:4041'
  }, cb);

  process.on('exit', function() {
    browserSync.exit();
  });

  gulp.watch('build/**/*.*', browserSync.reload);

  // FIX this part to only reload styles parts that changed
  // gulp.watch(['build/**/*.*'].concat(
  //   src.server.map(function(file) { return '!' + file; })
  // ), function(file) {
  //   browserSync.reload(path.relative(__dirname, file.path));
  // });
});

// Lint
gulp.task('lint', function() {
  return gulp.src(['app/**/*.js'])
    // eslint() attaches the lint output to the eslint property
    // of the file object so it can be used by other modules.
    .pipe($.eslint())
    // eslint.format() outputs the lint results to the console.
    // Alternatively use eslint.formatEach() (see Docs).
    .pipe($.eslint.format())
    // To have the process exit with an error code (1) on
    // lint error, return the stream and pipe to failOnError last.
    .pipe($.eslint.failOnError());
});
