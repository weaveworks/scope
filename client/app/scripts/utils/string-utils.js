import React from 'react';
import filesize from 'filesize';
import d3 from 'd3';

const formatLargeValue = d3.format('s');

function renderHtml(text, unit) {
  return (
    <span className="metric-formatted">
      <span className="metric-value">{text}</span>
      <span className="metric-unit">{unit}</span>
    </span>
  );
}


function makeFormatters(renderFn) {
  const formatters = {
    filesize(value) {
      const obj = filesize(value, {output: 'object'});
      return renderFn(obj.value, obj.suffix);
    },

    integer(value) {
      if (value < 1100 && value >= 0) {
        return Number(value).toFixed(0);
      }
      return formatLargeValue(value);
    },

    number(value) {
      if (value < 1100 && value >= 0) {
        return Number(value).toFixed(2);
      }
      return formatLargeValue(value);
    },

    percent(value) {
      return renderFn(formatters.number(value), '%');
    }
  };

  return formatters;
}


const formatters = makeFormatters(renderHtml);
const svgFormatters = makeFormatters((text, unit) => `${text}${unit}`);

export function formatMetric(value, opts, svg) {
  const formatterBase = svg ? svgFormatters : formatters;
  const formatter = opts && formatterBase[opts.format] ? opts.format : 'number';
  return formatterBase[formatter](value);
}

export const formatDate = d3.time.format.iso;
