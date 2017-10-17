import { fromJS } from 'immutable';


describe('MathUtils', () => {
  const MathUtils = require('../math-utils');

  describe('modulo', () => {
    const f = MathUtils.modulo;

    it('it should calculate the modulo (also for negatives)', () => {
      expect(f(5, 5)).toBe(0);
      expect(f(4, 5)).toBe(4);
      expect(f(3, 5)).toBe(3);
      expect(f(2, 5)).toBe(2);
      expect(f(1, 5)).toBe(1);
      expect(f(0, 5)).toBe(0);
      expect(f(-1, 5)).toBe(4);
      expect(f(-2, 5)).toBe(3);
      expect(f(-3, 5)).toBe(2);
      expect(f(-4, 5)).toBe(1);
      expect(f(-5, 5)).toBe(0);
    });
  });

  describe('minEuclideanDistanceBetweenPoints', () => {
    const f = MathUtils.minEuclideanDistanceBetweenPoints;
    const entryA = { pointA: { x: 0, y: 0 } };
    const entryB = { pointB: { x: 30, y: 0 } };
    const entryC = { pointC: { x: 0, y: -40 } };
    const entryD = { pointD: { x: -1000, y: 567 } };
    const entryE = { pointE: { x: -999, y: 567 } };
    const entryF = { pointF: { x: 30, y: 0 } };

    it('it should return the minimal distance between any two points in the collection', () => {
      expect(f(fromJS({}))).toBe(Infinity);
      expect(f(fromJS({...entryA}))).toBe(Infinity);
      expect(f(fromJS({...entryA, ...entryB}))).toBe(30);
      expect(f(fromJS({...entryA, ...entryC}))).toBe(40);
      expect(f(fromJS({...entryB, ...entryC}))).toBe(50);
      expect(f(fromJS({
        ...entryA, ...entryB, ...entryC, ...entryD
      }))).toBe(30);
      expect(f(fromJS({
        ...entryA, ...entryB, ...entryC, ...entryD, ...entryE
      }))).toBe(1);
      expect(f(fromJS({
        ...entryA, ...entryB, ...entryC, ...entryD, ...entryF
      }))).toBe(0);
    });
  });
});
