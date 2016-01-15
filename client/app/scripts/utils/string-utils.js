import React from 'react';
import filesize from 'filesize';

const formatters = {
  filesize(value) {
    const obj = filesize(value, {output: 'object'});
    return formatters.metric(obj.value, obj.suffix);
  },

  number(value) {
    return value;
  },

  percent(value) {
    return formatters.metric(value, '%');
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
