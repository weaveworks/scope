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
    // if (hsl.l > 0.5) {
    hsl = hsl.darker();
    // }

    return hsl.toString();
  }
};

module.exports = NodeColorMixin;
