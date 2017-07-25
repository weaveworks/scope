
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

function euclideanDistance(pointA, pointB) {
  const dx = pointA.get('x') - pointB.get('x');
  const dy = pointA.get('y') - pointB.get('y');
  return Math.sqrt((dx * dx) + (dy * dy));
}

// This could be solved in O(N log N) (see https://en.wikipedia.org/wiki/Closest_pair_of_points_problem),
// but this brute-force O(N^2) should be good enough for a reasonable number of nodes.
export function minEuclideanDistanceBetweenPoints(points) {
  let minDistance = Infinity;
  points.forEach((pointA, idA) => {
    points.forEach((pointB, idB) => {
      const distance = euclideanDistance(pointA, pointB);
      if (idA !== idB && distance < minDistance) {
        minDistance = distance;
      }
    });
  });
  return minDistance;
}
