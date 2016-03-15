jest.dontMock('../web-api-utils');

describe('WebApiUtils', () => {
  const WebApiUtils = require('../web-api-utils');

  describe('basePath', () => {
    const basePath = WebApiUtils.basePath;

    it('should handle /scope/terminal.html', () => {
      expect(basePath('/scope/terminal.html')).toBe('/scope');
    });

    it('should handle /scope/', () => {
      expect(basePath('/scope/')).toBe('/scope');
    });

    it('should handle /scope', () => {
      expect(basePath('/scope')).toBe('/scope');
    });

    it('should handle /', () => {
      expect(basePath('/')).toBe('');
    });
  });
});
