
import {OrderedMap as makeOrderedMap} from 'immutable';

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

  describe('buildOptionsQuery', () => {
    const buildOptionsQuery = WebApiUtils.buildOptionsQuery;

    it('should handle empty options', () => {
      expect(buildOptionsQuery(makeOrderedMap({}))).toBe('');
    });

    it('should combine multiple options', () => {
      expect(buildOptionsQuery(makeOrderedMap([
        ['foo', 2],
        ['bar', 4]
      ]))).toBe('foo=2&bar=4');
    });
  });
});
