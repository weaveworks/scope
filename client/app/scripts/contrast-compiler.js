/* eslint-disable class-methods-use-this */
// Webpack plugin for creating contrast mode stylesheet
const _ = require('lodash');

function findAsset(collection, name) {
  return _.find(collection, c => _.includes(c, name));
}

module.exports = class ContrastStyleCompiler {
  apply(compiler) {
    let themeJsChunk;

    compiler.plugin('compilation', (compilation) => {
      compilation.plugin('html-webpack-plugin-before-html-processing', (htmlPluginData, callback) => {
        themeJsChunk = findAsset(htmlPluginData.assets.js, 'contrast-theme');
        if (!themeJsChunk) {
          return callback(null, htmlPluginData);
        }
        // Find the name of the contrast stylesheet and save it to a window variable.
        const { css, publicPath } = htmlPluginData.assets;
        const contrast = findAsset(css, 'contrast-theme');
        const normal = findAsset(css, 'style-app');
        // Convert to JSON string so they can be parsed into a window variable
        const themes = JSON.stringify({ contrast, normal, publicPath });
        // Append a script to the end of <head /> to evaluate before the other scripts are loaded.
        const script = `<script>window.__WEAVE_SCOPE_THEMES = JSON.parse('${themes}')</script>`;
        const [head, end] = htmlPluginData.html.split('</head>');
        htmlPluginData.html = head.concat(script).concat('\n  </head>').concat(end);
        // Remove the contrast assets so they don't get added to the HTML.
        _.remove(htmlPluginData.assets.css, i => i === contrast);
        _.remove(htmlPluginData.assets.js, i => i === themeJsChunk);

        return callback(null, htmlPluginData);
      });
    });

    compiler.plugin('emit', (compilation, callback) => {
      // Remove the contrast-theme.js file, since it doesn't do anything
      const filename = themeJsChunk && themeJsChunk.split('?')[0];
      if (filename) {
        delete compilation.assets[filename];
      }
      callback();
    });
  }
};
