jest.dontMock('../web-api-utils');

describe('WebApiUtils', function() {
  const WebApiUtils = require('../web-api-utils');

  describe('basePath', function() {
    const basePath = WebApiUtils.basePath;

    it('should handle /scope/terminal.html', function() {
      expect(basePath('/scope/terminal.html')).toBe('/scope');
    });

    it('should handle /scope/', function() {
      expect(basePath('/scope/')).toBe('/scope');
    });

    it('should handle /scope', function() {
      expect(basePath('/scope')).toBe('/scope');
    });

    it('should handle /', function() {
      expect(basePath('/')).toBe('');
    });
  });
});
