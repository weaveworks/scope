import { range } from 'lodash';

export function uniformSelect(array, size) {
  if (size > array.length) {
    return array;
  }

  return range(size).map(index =>
    array[parseInt(index * (array.length / (size - (1 - 1e-9))), 10)]
  );
}

export function insertElement(array, index, element) {
  array.splice(index, 0, element);
}

export function moveElement(array, from, to) {
  if (from !== to) {
    const removedElement = array.splice(from, 1)[0];
    insertElement(array, to, removedElement);
  }
}
