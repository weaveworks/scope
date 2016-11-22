
describe('StringUtils', () => {
  const StringUtils = require('../string-utils');

  describe('formatMetric', () => {
    const f = StringUtils.formatMetric;

    it('it should render 0', () => {
      expect(f(0)).toBe('0.00');
    });
  });

  describe('longestCommonPrefix', () => {
    const f = StringUtils.longestCommonPrefix;

    it('it should return the longest common prefix', () => {
      expect(f(['interspecies', 'interstellar'])).toBe('inters');
      expect(f(['space', 'space'])).toBe('space');
      expect(f([''])).toBe('');
      expect(f(['prefix', 'suffix'])).toBe('');
    });
  });

  describe('ipToPaddedString', () => {
    const f = StringUtils.ipToPaddedString;

    it('it should return the formatted IP', () => {
      expect(f('10.244.253.4')).toBe('010.244.253.004');
      expect(f('0.24.3.4')).toBe('000.024.003.004');
    });
  });
});
