/* eslint-disable no-underscore-dangle */
import last from 'lodash/last';
/**
 * Change the Scope UI theme from normal to high-contrast.
 * This will inject a stylesheet into <head> and override the styles.
 *
 * A window-level variable is written to the .html page during the build process that contains
 * the filename (and content hash) needed to download the file.
 */

function getFilename(href) {
  return last(href.split('/'));
}

export function loadTheme(theme = 'normal') {
  if (window.__WEAVE_SCOPE_THEMES) {
    // Load the pre-built stylesheet.
    const stylesheet = window.__WEAVE_SCOPE_THEMES[theme];
    const head = document.querySelector('head');
    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = stylesheet;
    link.onload = () => {
      // Remove the old stylesheet to prevent weird overlapping styling issues
      const oldTheme = theme === 'normal' ? 'contrast' : 'normal';
      const links = document.querySelectorAll('head link');
      for (let i = 0; i < links.length; i += 1) {
        const l = links[i];
        if (getFilename(l.href) === getFilename(window.__WEAVE_SCOPE_THEMES[oldTheme])) {
          head.removeChild(l);
          break;
        }
      }
    };

    head.appendChild(link);
  }
}
