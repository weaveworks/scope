const d3 = require('d3');

const PSEUDO_COLOR = '#b1b1cb';
const hueRange = [20, 330]; // exclude red
const hueScale = d3.scale.linear().range(hueRange);
// map hues to lightness
const lightnessScale = d3.scale.linear().domain(hueRange).range([0.3, 0.7]);
const startLetterRange = 'A'.charCodeAt();
const endLetterRange = 'Z'.charCodeAt();
const letterRange = endLetterRange - startLetterRange;

/**
 * Converts a text to a 360 degree value
 */
function text2degree(text) {
  const input = text.substr(0, 2).toUpperCase();
  let num = 0;
  let charCode;
  for (let i = 0; i < input.length; i++) {
    charCode = Math.max(Math.min(input[i].charCodeAt(), endLetterRange), startLetterRange);
    num += Math.pow(letterRange, input.length - i - 1) * (charCode - startLetterRange);
  }
  hueScale.domain([0, Math.pow(letterRange, input.length)]);
  return hueScale(num);
}

function colors(text, secondText) {
  let hue = text2degree(text);
  // skip green and shift to the end of the color wheel
  if (hue > 70 && hue < 150) {
    hue += 80;
  }
  const saturation = 0.7;
  let lightness = 0.5;
  if (secondText) {
    // reuse text2degree and feed degree to lightness scale
    lightness = lightnessScale(text2degree(secondText));
  }
  const color = d3.hsl(hue, saturation, lightness);
  return color;
}

const NodeColorMixin = {
  getNodeColor: function(text, secondText) {
    return colors(text, secondText);
  },
  getNodeColorDark: function(text, secondText) {
    if (text === undefined) {
      return PSEUDO_COLOR;
    }
    const color = d3.rgb(colors(text, secondText));
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
