
describe('StringUtils', () => {
  const StringUtils = require('../string-utils');

  describe('formatMetric', () => {
    const formatMetric = StringUtils.formatMetric;

    it('it should render 0', () => {
      expect(formatMetric(0)).toBe('0.00');
    });
  });

  describe('longestCommonPrefix', () => {
    const fun = StringUtils.longestCommonPrefix;

    it('it should return the longest common prefix', () => {
      expect(fun(['interspecies', 'interstellar'])).toBe('inters');
      expect(fun(['space', 'space'])).toBe('space');
      expect(fun([''])).toBe('');
      expect(fun(['prefix', 'suffix'])).toBe('');
    });
  });

  describe('iPtoPaddedString', () => {
    const fun = StringUtils.iPtoPaddedString;

    it('it should return the formatted IP', () => {
      expect(fun('10.244.253.4')).toBe('010.244.253.004');
      expect(fun('0.24.3.4')).toBe('000.024.003.004');
    });
  });
});
