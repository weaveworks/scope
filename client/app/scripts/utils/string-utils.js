import React from 'react';
import filesize from 'filesize';
import d3 from 'd3';

const formatLargeValue = d3.format('s');

const formatters = {
  filesize(value) {
    const obj = filesize(value, {output: 'object'});
    return formatters.metric(obj.value, obj.suffix);
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
    return formatters.metric(formatters.number(value), '%');
  },

  metric(text, unit) {
    return (
      <span className="metric-formatted">
        <span className="metric-value">{text}</span>
        <span className="metric-unit">{unit}</span>
      </span>
    );
  }
};

export function formatMetric(value, opts) {
  const formatter = opts && formatters[opts.format] ? opts.format : 'number';
  return formatters[formatter](value);
}

export const formatDate = d3.time.format.iso;
