const { isString, kebabCase, forEach } = require('lodash');
const theme = require('weaveworks-ui-components/lib/theme').default;

// Flattens and collects all theme colors, names them
// as Scss vars and returns them as query string
// TODO: Move this helper to ui-components repo as
// it's currently used both here and in service-ui.
function themeColorsAsScss() {
  const colors = [];
  forEach(theme.colors, (value, name) => {
    const colorPrefix = `$color-${kebabCase(name)}`;
    if (isString(value)) {
      colors.push(`${colorPrefix}: ${value}`);
    } else {
      forEach(value, (innerValue, subname) => {
        colors.push(`${colorPrefix}-${kebabCase(subname)}: ${innerValue}`);
      });
    }
  });
  return `${colors.join('; ')};`;
}

module.exports = {
  themeColorsAsScss,
};
