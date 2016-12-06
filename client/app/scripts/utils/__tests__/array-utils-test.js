import _ from 'lodash';

describe('ArrayUtils', () => {
  const ArrayUtils = require('../array-utils');

  describe('uniformSelect', () => {
    const f = ArrayUtils.uniformSelect;

    it('it should select the array elements uniformly, including the endpoints', () => {
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
        const arr = _.range(1, 10001);
        expect(f(arr, 4)).toEqual([1, 3334, 6667, 10000]);
        expect(f(arr, 3)).toEqual([1, 5000, 10000]);
        expect(f(arr, 2)).toEqual([1, 10000]);
      }
    });
  });
});
