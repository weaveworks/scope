
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
});
