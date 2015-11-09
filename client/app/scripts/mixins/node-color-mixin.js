const d3 = require('d3');

const colors = d3.scale.category20();

// make sure the internet always gets the same color
const internetLabel = 'the Internet';
colors(internetLabel);


const NodeColorMixin = {
  getNodeColor: function(text) {
    return colors(text);
  },
  getNodeColorDark: function(text) {
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
