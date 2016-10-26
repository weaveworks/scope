
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
});
