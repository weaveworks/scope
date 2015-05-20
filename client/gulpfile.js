'use strict';

var gulp = require('gulp');
var connect = require('gulp-connect');
var livereload = require('gulp-livereload');
var browserify = require('browserify');
var del = require('del');
var source = require('vinyl-source-stream');
var buffer = require('vinyl-buffer');
var reactify = require('reactify');

// load plugins
var $ = require('gulp-load-plugins')();

var isDev = true;
var isProd = false;

gulp.task('styles', function () {
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
    var bundler = browserify('./app/scripts/main.js', {debug: isDev});
    bundler.transform(reactify);

    var stream = bundler.bundle();

    return stream
        .pipe(source('bundle.js'))
        .pipe($.if(isDev, gulp.dest('.tmp/scripts')))
        .pipe($.if(isProd, buffer()))
        .pipe($.if(isProd, $.uglify()))
        .pipe($.if(isProd, gulp.dest('dist/scripts')))
        .pipe(livereload())
        .on('error', $.util.log);
});

gulp.task('html', ['styles', 'scripts'], function () {
    return gulp.src('app/*.html')
        .pipe($.preprocess())
        .pipe(gulp.dest('dist'))
        .pipe($.size())
        .pipe(livereload());
});

gulp.task('images', function () {
    return gulp.src('app/images/**/*')
        .pipe(gulp.dest('dist/images'))
        .pipe($.size());
});

gulp.task('fonts', function () {
    return gulp.src('node_modules/font-awesome/fonts/*')
        .pipe($.filter('**/*.{eot,svg,ttf,woff,woff2}'))
        .pipe($.flatten())
        .pipe($.if(isDev, gulp.dest('.tmp/fonts')))
        .pipe($.if(isProd, gulp.dest('dist/fonts')))
        .pipe($.size());
});

gulp.task('extras', function () {
    return gulp.src(['app/*.*', '!app/*.html'], { dot: true })
        .pipe(gulp.dest('dist'));
});

gulp.task('clean', function () {
    return del(['.tmp', 'dist']);
});

gulp.task('production', ['html', 'images', 'fonts', 'extras']);

gulp.task('build', function () {
    isDev = false;
    isProd = true;
    gulp.start('production');
});

gulp.task('default', ['clean'], function () {
    gulp.start('build');
});

gulp.task('connect', function () {
    connect.server({
        root: ['.tmp', 'app'],
        port: 4041,
        middleware: function(connect, o) {
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

gulp.task('serve', ['connect', 'styles', 'scripts', 'fonts']);

gulp.task('watch', ['serve'], function () {
    livereload.listen();
    gulp.watch('app/styles/**/*.less', ['styles']);
    gulp.watch('app/scripts/**/*.js', ['scripts']);
    gulp.watch('app/images/**/*', ['images']);
});
