
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
