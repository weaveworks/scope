jest.dontMock('../string-utils');

describe('StringUtils', function() {
  const StringUtils = require('../string-utils');

  describe('formatMetric', function() {
    const formatMetric = StringUtils.formatMetric;

    it('it should render 0', function() {
      expect(formatMetric(0)).toBe('0.00');
    });
  });
});
