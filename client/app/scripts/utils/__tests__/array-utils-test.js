import _ from 'lodash';

describe('ArrayUtils', () => {
  const ArrayUtils = require('../array-utils');

  describe('uniformSelect', () => {
    const f = ArrayUtils.uniformSelect;

    it('it should select the array elements uniformly, including the endpoints', () => {
      expect(f(['x', 'y'], 3)).toEqual(['x', 'y']);
      expect(f(['x', 'y'], 2)).toEqual(['x', 'y']);

      expect(f(['A', 'B', 'C', 'D', 'E'], 6)).toEqual(['A', 'B', 'C', 'D', 'E']);
      expect(f(['A', 'B', 'C', 'D', 'E'], 5)).toEqual(['A', 'B', 'C', 'D', 'E']);
      expect(f(['A', 'B', 'C', 'D', 'E'], 4)).toEqual(['A', 'B', 'D', 'E']);
      expect(f(['A', 'B', 'C', 'D', 'E'], 3)).toEqual(['A', 'C', 'E']);
      expect(f(['A', 'B', 'C', 'D', 'E'], 2)).toEqual(['A', 'E']);

      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 12)).toEqual(
        [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
      );
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 11)).toEqual(
        [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]);
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 10)).toEqual([1, 2, 3, 4, 5, 7, 8, 9, 10, 11]
      );
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 9)).toEqual([1, 2, 3, 5, 6, 7, 9, 10, 11]);
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 8)).toEqual([1, 2, 4, 5, 7, 8, 10, 11]);
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 7)).toEqual([1, 2, 4, 6, 8, 10, 11]);
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 6)).toEqual([1, 3, 5, 7, 9, 11]);
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 5)).toEqual([1, 3, 6, 9, 11]);
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 4)).toEqual([1, 4, 8, 11]);
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 3)).toEqual([1, 6, 11]);
      expect(f([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], 2)).toEqual([1, 11]);

      expect(f(_.range(1, 10001), 4)).toEqual([1, 3334, 6667, 10000]);
      expect(f(_.range(1, 10001), 3)).toEqual([1, 5000, 10000]);
      expect(f(_.range(1, 10001), 2)).toEqual([1, 10000]);
    });
  });
});
