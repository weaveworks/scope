import moment from 'moment';

// Replacement for timely dependency
export function timer(fn) {
  const timedFn = (...args) => {
    const start = new Date();
    const result = fn.apply(fn, args);
    timedFn.time = new Date() - start;
    return result;
  };
  return timedFn;
}

export function nowInSecondsPrecision() {
  return moment().startOf('second');
}

export function clampToNowInSecondsPrecision(timestamp) {
  const now = nowInSecondsPrecision();
  return timestamp.isAfter(now) ? now : timestamp;
}

// This is unfortunately not there in moment.js
export function scaleDuration(duration, scale) {
  return moment.duration(duration.asMilliseconds() * scale);
}

export function timestampsEqual(timestampA, timestampB) {
  const stringifiedTimestampA = timestampA ? timestampA.toISOString() : '';
  const stringifiedTimestampB = timestampB ? timestampB.toISOString() : '';
  return stringifiedTimestampA === stringifiedTimestampB;
}
