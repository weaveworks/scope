import { range } from 'lodash';

// NOTE: All the array operations defined here should be non-mutating.

export function uniformSelect(array, size) {
  if (size > array.length) {
    return array;
  }

  return range(size).map(index =>
    array[parseInt(index * (array.length / (size - (1 - 1e-9))), 10)]
  );
}

export function insertElement(array, index, element) {
  return array.slice(0, index).concat([element], array.slice(index));
}

export function removeElement(array, index) {
  return array.slice(0, index).concat(array.slice(index + 1));
}

export function moveElement(array, from, to) {
  if (from === to) {
    return array;
  }
  return insertElement(removeElement(array, from), to, array[from]);
}
