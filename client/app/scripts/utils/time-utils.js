
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

export function timestampsEqual(timestampA, timestampB) {
  const stringifiedTimestampA = timestampA ? timestampA.toISOString() : '';
  const stringifiedTimestampB = timestampB ? timestampB.toISOString() : '';
  return stringifiedTimestampA === stringifiedTimestampB;
}
