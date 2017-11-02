import moment from 'moment';


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

  describe('humanizedRoundedDownDuration', () => {
    const f = StringUtils.humanizedRoundedDownDuration;

    it('it should return the humanized duration', () => {
      expect(f(moment.duration(0))).toBe('now');
      expect(f(moment.duration(0.9 * 1000))).toBe('now');
      expect(f(moment.duration(1 * 1000))).toBe('1 second');
      expect(f(moment.duration(8.62 * 60 * 1000))).toBe('8 minutes');
      expect(f(moment.duration(14.99 * 60 * 60 * 1000))).toBe('14 hours');
      expect(f(moment.duration(5.2 * 24 * 60 * 60 * 1000))).toBe('5 days');
      expect(f(moment.duration(11.8 * 30 * 24 * 60 * 60 * 1000))).toBe('11 months');
      expect(f(moment.duration(12.8 * 30 * 24 * 60 * 60 * 1000))).toBe('1 year');
      expect(f(moment.duration(9.4 * 12 * 30 * 24 * 60 * 60 * 1000))).toBe('9 years');
    });
  });
});
