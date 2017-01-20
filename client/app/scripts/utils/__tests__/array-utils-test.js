import { range } from 'lodash';

function testNotMutatingArray(f, array, ...otherArgs) {
  const original = array.slice();
  f(array, ...otherArgs);
  expect(array).toEqual(original);
}

describe('ArrayUtils', () => {
  const ArrayUtils = require('../array-utils');

  describe('uniformSelect', () => {
    const f = ArrayUtils.uniformSelect;

    it('it should select the array elements uniformly, including the endpoints', () => {
      testNotMutatingArray(f, ['A', 'B', 'C', 'D', 'E'], 3);
      {
        const arr = ['x', 'y'];
        expect(f(arr, 3)).toEqual(['x', 'y']);
        expect(f(arr, 2)).toEqual(['x', 'y']);
      }
      {
        const arr = ['A', 'B', 'C', 'D', 'E'];
        expect(f(arr, 6)).toEqual(['A', 'B', 'C', 'D', 'E']);
        expect(f(arr, 5)).toEqual(['A', 'B', 'C', 'D', 'E']);
        expect(f(arr, 4)).toEqual(['A', 'B', 'D', 'E']);
        expect(f(arr, 3)).toEqual(['A', 'C', 'E']);
        expect(f(arr, 2)).toEqual(['A', 'E']);
      }
      {
        const arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11];
        expect(f(arr, 12)).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]);
        expect(f(arr, 11)).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]);
        expect(f(arr, 10)).toEqual([1, 2, 3, 4, 5, 7, 8, 9, 10, 11]);
        expect(f(arr, 9)).toEqual([1, 2, 3, 5, 6, 7, 9, 10, 11]);
        expect(f(arr, 8)).toEqual([1, 2, 4, 5, 7, 8, 10, 11]);
        expect(f(arr, 7)).toEqual([1, 2, 4, 6, 8, 10, 11]);
        expect(f(arr, 6)).toEqual([1, 3, 5, 7, 9, 11]);
        expect(f(arr, 5)).toEqual([1, 3, 6, 9, 11]);
        expect(f(arr, 4)).toEqual([1, 4, 8, 11]);
        expect(f(arr, 3)).toEqual([1, 6, 11]);
        expect(f(arr, 2)).toEqual([1, 11]);
      }
      {
        const arr = range(1, 10001);
        expect(f(arr, 4)).toEqual([1, 3334, 6667, 10000]);
        expect(f(arr, 3)).toEqual([1, 5000, 10000]);
        expect(f(arr, 2)).toEqual([1, 10000]);
      }
    });
  });

  describe('insertElement', () => {
    const f = ArrayUtils.insertElement;

    it('it should insert an element into the array at the specified index', () => {
      testNotMutatingArray(f, ['x', 'y', 'z'], 0, 'a');
      expect(f(['x', 'y', 'z'], 0, 'a')).toEqual(['a', 'x', 'y', 'z']);
      expect(f(['x', 'y', 'z'], 1, 'a')).toEqual(['x', 'a', 'y', 'z']);
      expect(f(['x', 'y', 'z'], 2, 'a')).toEqual(['x', 'y', 'a', 'z']);
      expect(f(['x', 'y', 'z'], 3, 'a')).toEqual(['x', 'y', 'z', 'a']);
    });
  });

  describe('removeElement', () => {
    const f = ArrayUtils.removeElement;

    it('it should remove the element at the specified index from the array', () => {
      testNotMutatingArray(f, ['x', 'y', 'z'], 0);
      expect(f(['x', 'y', 'z'], 0)).toEqual(['y', 'z']);
      expect(f(['x', 'y', 'z'], 1)).toEqual(['x', 'z']);
      expect(f(['x', 'y', 'z'], 2)).toEqual(['x', 'y']);
    });
  });

  describe('moveElement', () => {
    const f = ArrayUtils.moveElement;

    it('it should move an array element, modifying the array', () => {
      testNotMutatingArray(f, ['x', 'y', 'z'], 0, 1);
      expect(f(['x', 'y', 'z'], 0, 1)).toEqual(['y', 'x', 'z']);
      expect(f(['x', 'y', 'z'], 1, 0)).toEqual(['y', 'x', 'z']);
      expect(f(['x', 'y', 'z'], 0, 2)).toEqual(['y', 'z', 'x']);
      expect(f(['x', 'y', 'z'], 2, 0)).toEqual(['z', 'x', 'y']);
      expect(f(['x', 'y', 'z'], 1, 2)).toEqual(['x', 'z', 'y']);
      expect(f(['x', 'y', 'z'], 2, 1)).toEqual(['x', 'z', 'y']);
      expect(f(['x', 'y', 'z'], 0, 0)).toEqual(['x', 'y', 'z']);
      expect(f(['x', 'y', 'z'], 1, 1)).toEqual(['x', 'y', 'z']);
      expect(f(['x', 'y', 'z'], 2, 2)).toEqual(['x', 'y', 'z']);
      expect(f(['a', 'b', 'c', 'd', 'e'], 4, 1)).toEqual(['a', 'e', 'b', 'c', 'd']);
      expect(f(['a', 'b', 'c', 'd', 'e'], 1, 4)).toEqual(['a', 'c', 'd', 'e', 'b']);
      expect(f(['a', 'b', 'c', 'd', 'e'], 1, 3)).toEqual(['a', 'c', 'd', 'b', 'e']);
    });
  });
});
