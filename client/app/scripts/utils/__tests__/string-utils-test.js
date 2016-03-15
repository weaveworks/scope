jest.dontMock('../string-utils');

describe('StringUtils', () => {
  const StringUtils = require('../string-utils');

  describe('formatMetric', () => {
    const formatMetric = StringUtils.formatMetric;

    it('it should render 0', () => {
      expect(formatMetric(0)).toBe('0.00');
    });
  });
});
