import React from 'react';
import filesize from 'filesize';
import { format as d3Format } from 'd3-format';
import { isoFormat } from 'd3-time-format';
import LCP from 'lcp';
import moment from 'moment';

import { round } from './math-utils';

const formatLargeValue = d3Format('s');
const formatFlexiblePrecision = v => round(v, 4).toString();

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


function padToThreeDigits(n) {
  return `000${n}`.slice(-3);
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

    percent(value, { fixedPrecision = true }) {
      const format = fixedPrecision ? formatters.number : formatFlexiblePrecision;
      return renderFn(format(value), '%');
    }
  };

  return formatters;
}


function makeFormatMetric(renderFn) {
  const formatters = makeFormatters(renderFn);
  return (value, opts) => {
    const formatter = opts && formatters[opts.format] ? opts.format : 'number';
    return formatters[formatter](value, opts);
  };
}


export const formatMetric = makeFormatMetric(renderHtml);
export const formatMetricSvg = makeFormatMetric(renderSvg);
export const formatDate = isoFormat; // d3.time.format.iso;

const CLEAN_LABEL_REGEX = /[^A-Za-z0-9]/g;
export function slugify(label) {
  return label.replace(CLEAN_LABEL_REGEX, '').toLowerCase();
}

export function longestCommonPrefix(strArr) {
  return (new LCP(strArr)).lcp();
}

// Converts IPs from '10.244.253.4' to '010.244.253.004' format.
export function ipToPaddedString(value) {
  return value.match(/\d+/g).map(padToThreeDigits).join('.');
}

// Formats metadata values. Add a key to the `formatters` obj
// that matches the `dataType` of the field. You must return an Object
// with the keys `value` and `title` defined.
export function formatDataType(field) {
  const formatters = {
    datetime(dateString) {
      const date = moment(new Date(dateString));
      return {
        value: date.fromNow(),
        title: date.format('YYYY-MM-DD HH:mm:ss.SSS')
      };
    }
  };
  const format = formatters[field.dataType];
  return format
    ? format(field.value)
    : {value: field.value, title: field.value};
}
