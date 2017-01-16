
export function waterfall(series, target, cb) {
  function next(result) {
    const fn = series.shift();
    if (fn) {
      try {
        fn(result, next);
      } catch (e) {
        cb(e);
      }
    } else {
      cb(null, result);
    }
  }
  next(target, next);
}
