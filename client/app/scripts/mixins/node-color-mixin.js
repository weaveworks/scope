var d3 = require('d3');

var colors = d3.scale.category20();

// make sure the internet always gets the same color
var internetLabel = "the Internet";
colors(internetLabel);


var NodeColorMixin = {
  getNodeColor: function(text) {
    return colors(text);
  },
  getNodeColorDark: function(text) {
    var color = d3.rgb(colors(text));
    var hsl = color.hsl();

    // ensure darkness
    // if (hsl.l > 0.5) {
      hsl = hsl.darker();
    // }

    return hsl.toString();
  }
};

module.exports = NodeColorMixin;