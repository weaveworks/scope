import { range } from 'lodash';

export function uniformSelect(array, size) {
  if (size > array.length) {
    return array;
  }

  return range(size).map(index =>
    array[parseInt(index * (array.length / (size - (1 - 1e-9))), 10)]
  );
}
