import React from 'react';
import filesize from 'filesize';
import { format as d3Format } from 'd3-format';
import { isoFormat } from 'd3-time-format';
import LCP from 'lcp';
import moment from 'moment';

const formatLargeValue = d3Format('s');


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

// Doing the manual parsing because `duration.humanize()` would sometimes round up the period,
// while we always want a rounded down value for consistency with other values sent by backend.
export function humanizedRoundedDownDuration(duration) {
  let humanizedDuration = 'now';
  ['second', 'minute', 'hour', 'day', 'month', 'year'].forEach((period) => {
    const durationAsPeriod = Math.floor(duration.as(period));
    if (durationAsPeriod > 0) {
      const pluralEnding = ((durationAsPeriod !== 11 && (durationAsPeriod % 10) === 1) ? '' : 's');
      humanizedDuration = `${durationAsPeriod} ${period}${pluralEnding}`;
    }
  });
  return humanizedDuration;
}

// Formats metadata values. Add a key to the `formatters` obj
// that matches the `dataType` of the field. You must return an Object
// with the keys `value` and `title` defined.
// `referenceTimestamp` is the timestamp we've time-travelled to.
export function formatDataType(field, referenceTimestamp = null) {
  const formatters = {
    datetime(timestampString) {
      const timestamp = moment(timestampString);
      return {
        title: timestamp.utc().toISOString(),
        value: timestamp.from(referenceTimestamp ? moment(referenceTimestamp) : moment())
      };
    },
    duration(durationSecondsString) {
      const duration = moment.duration(Number(durationSecondsString), 'seconds');
      const humanizedDuration = humanizedRoundedDownDuration(duration);

      return {
        title: humanizedDuration,
        value: humanizedDuration,
      };
    },
  };
  const format = formatters[field.dataType];
  return format
    ? format(field.value)
    : { title: field.value, value: field.value };
}
