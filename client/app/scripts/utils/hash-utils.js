import { isPlainObject, mapValues, isEmpty, omitBy } from 'lodash';


export function hashDifferenceDeep(A, B) {
  // If the elements have exactly the same content, the difference is an empty object.
  // This could fail if the objects are both hashes with different permutation of keys,
  // but this case we handle below by digging in recursively.
  if (JSON.stringify(A) === JSON.stringify(B)) return {};

  // Otherwise, if either element is not a hash, always return the first element
  // unchanged as this function only takes difference of hash objects.
  if (!isPlainObject(A) || !isPlainObject(B)) return A;

  // If both elements are hashes, recursively take the difference by all keys
  const rawDiff = mapValues(A, (value, key) => hashDifferenceDeep(value, B[key]));

  // ... and filter out all the empty values.
  return omitBy(rawDiff, value => isEmpty(value));
}
