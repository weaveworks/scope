
const gulp = require('gulp');
const connect = require('gulp-connect');
const livereload = require('gulp-livereload');
const babelify = require('babelify');
const browserify = require('browserify');
const del = require('del');
const source = require('vinyl-source-stream');
const vbuffer = require('vinyl-buffer');
const reactify = require('reactify');

// load plugins
const $ = require('gulp-load-plugins')();

var isDev = true;
var isProd = false;

gulp.task('styles', function() {
  return gulp.src('app/styles/main.less')
    .pipe($.if(isDev, $.sourcemaps.init()))
    .pipe($.less())
    .pipe($.autoprefixer('last 1 version'))
    .pipe($.if(isDev, $.sourcemaps.write()))
    .pipe($.if(isDev, gulp.dest('.tmp/styles')))
    .pipe($.if(isProd, $.csso()))
    .pipe($.if(isProd, gulp.dest('dist/styles')))
    .pipe($.size())
    .pipe(livereload());
});

gulp.task('scripts', function() {
  const bundler = browserify('./app/scripts/main.js', {debug: isDev});
  bundler.transform(reactify);
  bundler.transform(babelify);

  const stream = bundler.bundle();

  return stream
    .pipe(source('bundle.js'))
    .pipe($.if(isDev, gulp.dest('.tmp/scripts')))
    .pipe($.if(isProd, vbuffer()))
    .pipe($.if(isProd, $.uglify()))
    .on('error', $.util.log)
    .pipe($.if(isProd, gulp.dest('dist/scripts')))
    .pipe(livereload());
});

gulp.task('html', ['styles', 'scripts'], function() {
  return gulp.src('app/*.html')
    .pipe($.preprocess())
    .pipe(gulp.dest('dist'))
    .pipe($.size())
    .pipe(livereload());
});

gulp.task('images', function() {
  return gulp.src('app/images/**/*')
    .pipe(gulp.dest('dist/images'))
    .pipe($.size());
});

gulp.task('fonts', function() {
  return gulp.src('node_modules/font-awesome/fonts/*')
    .pipe($.filter('**/*.{eot,svg,ttf,woff,woff2}'))
    .pipe($.flatten())
    .pipe($.if(isDev, gulp.dest('.tmp/fonts')))
    .pipe($.if(isProd, gulp.dest('dist/fonts')))
    .pipe($.size());
});

gulp.task('extras', function() {
  return gulp.src(['app/*.*', '!app/*.html'], { dot: true })
    .pipe(gulp.dest('dist'));
});

gulp.task('clean', function() {
  return del(['.tmp', 'dist']);
});


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

gulp.task('production', ['html', 'images', 'fonts', 'extras']);

gulp.task('build', function() {
  isDev = false;
  isProd = true;
  gulp.start('production');
});

gulp.task('default', ['clean'], function() {
  gulp.start('build');
});

gulp.task('connect', function() {
  const root = isProd ? ['dist'] : ['.tmp', 'app'];
  connect.server({
    root: root,
    port: 4041,
    middleware: function() {
      return [(function() {
        const url = require('url');
        const proxy = require('proxy-middleware');
        const options = url.parse('http://localhost:4040/api');
        options.route = '/api';
        return proxy(options);
      })()];
    },
    livereload: false
  });
});

gulp.task('serve', ['connect', 'styles', 'scripts', 'fonts']);

gulp.task('serve-build', function() {
  isDev = false;
  isProd = true;

  // use local WS api
  gulp.src('app/*.html')
    .pipe($.preprocess({context: {DEBUG: true} }))
    .pipe(gulp.dest('dist'));

  gulp.start('connect');
});

gulp.task('watch', ['serve'], function() {
  livereload.listen();
  gulp.watch('app/styles/**/*.less', ['styles']);
  gulp.watch('app/scripts/**/*.js', ['scripts']);
  gulp.watch('app/images/**/*', ['images']);
});
