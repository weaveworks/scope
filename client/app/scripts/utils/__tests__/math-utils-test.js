
describe('MathUtils', () => {
  const MathUtils = require('../math-utils');

  describe('module', () => {
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

  describe('round', () => {
    const f = MathUtils.round;

    it('it should round the decimal number to given precision', () => {
      expect(f(-173.6499023, -2)).toBe(-200);
      expect(f(-173.6499023, -1)).toBe(-170);
      expect(f(-173.6499023, 0)).toBe(-174);
      expect(f(-173.6499023)).toBe(-174);
      expect(f(-173.6499023, 1)).toBe(-173.6);
      expect(f(-173.6499023, 2)).toBe(-173.65);
      expect(f(0.0013, 2)).toBe(0);
      expect(f(0.0013, 3)).toBe(0.001);
      expect(f(0.0013, 4)).toBe(0.0013);
      expect(f(0.0013, 5)).toBe(0.0013);
    });
  });
});
