
// http://stackoverflow.com/questions/4467539/javascript-modulo-not-behaving
//
// A modulo that "behaves" w/ negatives.
//
// modulo(5, 5) => 0
// modulo(4, 5) => 4
// modulo(3, 5) => 3
// modulo(2, 5) => 2
// modulo(1, 5) => 1
// modulo(0, 5) => 0
// modulo(-1, 5) => 4
// modulo(-2, 5) => 3
// modulo(-3, 5) => 2
// modulo(-4, 5) => 1
// modulo(-5, 5) => 0
//
export function modulo(i, n) {
  return ((i % n) + n) % n;
}

// Does the same that the deprecated d3.round was doing.
// Possibly imprecise: This https://github.com/d3/d3/issues/210
export function round(value, decimals = 0) {
  const p = Math.pow(10, decimals);
  return Math.round(value * p) / p;
}
