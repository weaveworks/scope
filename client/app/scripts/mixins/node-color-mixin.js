const d3 = require('d3');

const colors = d3.scale.category20();
const PSEUDO_COLOR = '#b1b1cb';
let colorIndex = 0;

const NodeColorMixin = {
  getNodeColor: function(text) {
    colorIndex++;
    // skip green and red (index 5-8 in d3.scale.category20)
    if (colorIndex > 4 && colorIndex < 9) {
      colors('_' + colorIndex);
      return this.getNodeColor(text);
    }

    return colors(text);
  },
  getNodeColorDark: function(text) {
    if (text === undefined) {
      return PSEUDO_COLOR;
    }
    const color = d3.rgb(colors(text));
    let hsl = color.hsl();

    // ensure darkness
    if (hsl.l > 0.7) {
      hsl = hsl.darker(1.5);
    } else {
      hsl = hsl.darker(1);
    }

    return hsl.toString();
  },
  brightenColor: function(color) {
    let hsl = d3.rgb(color).hsl();
    if (hsl.l > 0.5) {
      hsl = hsl.brighter(0.5);
    } else {
      hsl = hsl.brighter(0.8);
    }
    return hsl.toString();
  }
};

module.exports = NodeColorMixin;
