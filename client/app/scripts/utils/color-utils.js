import { hsl } from 'd3-color';
import { scaleLinear, scaleOrdinal } from 'd3-scale';
import { schemeCategory10 } from 'd3-scale-chromatic';

const PSEUDO_COLOR = '#b1b1cb';
const hueRange = [20, 330]; // exclude red
const hueScale = scaleLinear().range(hueRange);
const networkColorScale = scaleOrdinal(schemeCategory10);
// map hues to lightness
const lightnessScale = scaleLinear().domain(hueRange).range([0.5, 0.7]);
const startLetterRange = 'A'.charCodeAt();
const endLetterRange = 'Z'.charCodeAt();
const letterRange = endLetterRange - startLetterRange;

/**
 * Converts a text to a 360 degree value
 */
export function text2degree(text) {
  const input = text.substr(0, 2).toUpperCase();
  let num = 0;
  for (let i = 0; i < input.length; i += 1) {
    const charCode = Math.max(Math.min(input[i].charCodeAt(), endLetterRange), startLetterRange);
    num += Math.pow(letterRange, input.length - i - 1) * (charCode - startLetterRange);
  }
  hueScale.domain([0, Math.pow(letterRange, input.length)]);
  return hueScale(num);
}

export function colors(text, secondText) {
  let hue = text2degree(text);
  // skip green and shift to the end of the color wheel
  if (hue > 70 && hue < 150) {
    hue += 80;
  }
  const saturation = 0.6;
  let lightness = 0.5;
  if (secondText) {
    // reuse text2degree and feed degree to lightness scale
    lightness = lightnessScale(text2degree(secondText));
  }
  return hsl(hue, saturation, lightness);
}

export function getNeutralColor() {
  return PSEUDO_COLOR;
}

export function getNodeColor(text = '', secondText = '', isPseudo = false) {
  if (isPseudo) {
    return PSEUDO_COLOR;
  }
  return colors(text, secondText).toString();
}

export function getNodeColorDark(text = '', secondText = '', isPseudo = false) {
  if (isPseudo) {
    return PSEUDO_COLOR;
  }
  let color = hsl(colors(text, secondText));

  // ensure darkness
  if (color.h > 20 && color.h < 120) {
    color = color.darker(2);
  } else if (hsl.l > 0.7) {
    color = color.darker(1.5);
  } else {
    color = color.darker(1);
  }

  return color.toString();
}

export function getNetworkColor(text) {
  return networkColorScale(text);
}

export function brightenColor(c) {
  let color = hsl(c);
  if (hsl.l > 0.5) {
    color = color.brighter(0.5);
  } else {
    color = color.brighter(0.8);
  }
  return color.toString();
}

export function darkenColor(c) {
  let color = hsl(c);
  if (hsl.l < 0.5) {
    color = color.darker(0.5);
  } else {
    color = color.darker(0.8);
  }
  return color.toString();
}
