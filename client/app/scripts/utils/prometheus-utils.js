import reqwest from 'reqwest';

import { INTERVAL_SECS, QUERY_WINDOW_SECS } from '../constants/timer';

// Turn a specification into a Prometheus label match expression
function matchExpr(spec) {
  const labels = [];
  Object.keys(spec).forEach(k => {
    let eq = '=';
    let val = spec[k];
    if (Array.isArray(val)) {
      eq = '=~';
      val = val.join('|');
    }
    labels.push(k + eq + '"' + val + '"');
  });
  return labels.join(',');
}

export function processValues(val) {
  return {
    date: new Date(Math.round(val[0]) * 1000),
    value: parseFloat(val[1])
  };
}

export function requestRange(spec, start, end, success) {
  const query = 'query=sum(rate(flux_http_total{' + matchExpr(spec) + '}[' + INTERVAL_SECS + 's])) by (code)';
  const interval = '&step=' + INTERVAL_SECS + 's&start=' + start + '&end=' + end;
  const url = '/stats/api/v1/query_range?' + query + interval;

  reqwest({url, success});
}

function processLastValues(result) {
  return result.map(obj => [obj.metric.individual, parseFloat(obj.value[1])]);
}

export function requestLastValues(instances, cb) {
  const spec = {individual: instances};
  const query = 'query=sum(rate(flux_http_total{' + matchExpr(spec) + '}[' + QUERY_WINDOW_SECS + 's])) by (individual)';
  const url = '/stats/api/v1/query?' + query;
  const success = json => cb(processLastValues(json.data.result));
  reqwest({url, success});
}

export function now() {
  return Math.floor(+new Date / 1000 / INTERVAL_SECS) * INTERVAL_SECS;
}
