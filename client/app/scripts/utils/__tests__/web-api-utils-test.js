import { Map as makeMap, OrderedMap as makeOrderedMap } from 'immutable';

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
    let state = makeMap();

    it('should handle empty options', () => {
      expect(buildUrlQuery(makeOrderedMap([]), state)).toBe('');
    });

    it('should combine multiple options', () => {
      expect(buildUrlQuery(makeOrderedMap([
        ['foo', 2],
        ['bar', 4]
      ]), state)).toBe('foo=2&bar=4');
    });

    it('should combine multiple options with a timestamp', () => {
      state = state.set('pausedAt', '2015-06-14T21:12:05.275Z');
      expect(buildUrlQuery(makeOrderedMap([
        ['foo', 2],
        ['bar', 4]
      ]), state)).toBe('foo=2&bar=4&timestamp=2015-06-14T21:12:05.275Z');
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
      // instance ID first to match Weave Cloud routes
      expect(getApiPath('/proud-cloud-77/app')).toEqual('/api/app/proud-cloud-77');
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
      expect(getWebsocketUrl(host, '/proud-cloud-77/app')).toEqual(`ws://${host}/api/app/proud-cloud-77`);
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
