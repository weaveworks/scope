
import {OrderedMap as makeOrderedMap} from 'immutable';
import { buildUrlQuery, basePath, getApiPath, getWebsocketUrl } from '../web-api-utils';

describe('WebApiUtils', () => {
  describe('basePath', () => {
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

  describe('buildUrlQuery', () => {
    it('should handle empty options', () => {
      expect(buildUrlQuery(makeOrderedMap({}))).toBe('');
    });

    it('should combine multiple options', () => {
      expect(buildUrlQuery(makeOrderedMap([
        ['foo', 2],
        ['bar', 4]
      ]))).toBe('foo=2&bar=4');
    });
  });

  describe('getApiPath', () => {
    afterEach(() => {
      delete process.env.SCOPE_API_PREFIX;
    });
    it('returns the correct url when running standalone', () => {
      expect(getApiPath('/')).toEqual('');
    });
    it('returns the correct url when running in an iframe', () => {
      expect(getApiPath('/api/app/proud-cloud-77')).toEqual('/api/app/proud-cloud-77');
    });
    it('returns the correct url when running as a component', () => {
      process.env.SCOPE_API_PREFIX = '/api';
      expect(getApiPath('/app/proud-cloud-77')).toEqual('/api/app/proud-cloud-77');
    });
    it('returns the correct url from an arbitrary path', () => {
      expect(getApiPath('/demo/')).toEqual('/demo');
    });
    it('returns the correct url from an *.html page', () => {
      expect(getApiPath('/contrast.html')).toEqual('');
    });
    it('returns the correct url from an /*.html page while in an iframe', () => {
      expect(getApiPath('/api/app/proud-cloud-77/contrast.html')).toEqual('/api/app/proud-cloud-77');
    });
  });

  describe('getWebsocketUrl', () => {
    const host = 'localhost:4042';
    afterEach(() => {
      delete process.env.SCOPE_API_PREFIX;
    });
    it('returns the correct url when running standalone', () => {
      expect(getWebsocketUrl(host, '/')).toEqual(`ws://${host}`);
    });
    it('returns the correct url when running in an iframe', () => {
      expect(getWebsocketUrl(host, '/api/app/proud-cloud-77')).toEqual(`ws://${host}/api/app/proud-cloud-77`);
    });
    it('returns the correct url when running as a component', () => {
      process.env.SCOPE_API_PREFIX = '/api';
      expect(getWebsocketUrl(host, '/app/proud-cloud-77')).toEqual(`ws://${host}/api/app/proud-cloud-77`);
    });
    it('returns the correct url from an arbitrary path', () => {
      expect(getWebsocketUrl(host, '/demo/')).toEqual(`ws://${host}/demo`);
    });
    it('returns the correct url from an *.html page', () => {
      expect(getWebsocketUrl(host, '/contrast.html')).toEqual(`ws://${host}`);
    });
    it('returns the correct url from an /*.html page while in an iframe', () => {
      expect(getWebsocketUrl(host, '/api/app/proud-cloud-77/contrast.html')).toEqual(`ws://${host}/api/app/proud-cloud-77`);
    });
  });
});
