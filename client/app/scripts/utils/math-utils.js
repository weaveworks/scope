
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
