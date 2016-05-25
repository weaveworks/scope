import React from 'react';
import filesize from 'filesize';
import d3 from 'd3';
import LCP from 'lcp';

const formatLargeValue = d3.format('s');


function renderHtml(text, unit) {
  return (
    <span className="metric-formatted">
      <span className="metric-value">{text}</span>
      <span className="metric-unit">{unit}</span>
    </span>
  );
}


function renderSvg(text, unit) {
  return `${text}${unit}`;
}


function makeFormatters(renderFn) {
  const formatters = {
    filesize(value) {
      const obj = filesize(value, {output: 'object', round: 1});
      return renderFn(obj.value, obj.suffix);
    },

    integer(value) {
      const intNumber = Number(value).toFixed(0);
      if (value < 1100 && value >= 0) {
        return intNumber;
      }
      return formatLargeValue(intNumber);
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


function makeFormatMetric(renderFn) {
  const formatters = makeFormatters(renderFn);
  return (value, opts) => {
    const formatter = opts && formatters[opts.format] ? opts.format : 'number';
    return formatters[formatter](value);
  };
}


export const formatMetric = makeFormatMetric(renderHtml);
export const formatMetricSvg = makeFormatMetric(renderSvg);
export const formatDate = d3.time.format.iso;

const CLEAN_LABEL_REGEX = /\W/g;
export function slugify(label) {
  return label.replace(CLEAN_LABEL_REGEX, '').toLowerCase();
}

export function longestCommonPrefix(strArr) {
  return (new LCP(strArr)).lcp();
}
