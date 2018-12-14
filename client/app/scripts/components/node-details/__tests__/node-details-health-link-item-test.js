import moment from 'moment';

import { appendTime } from '../node-details-health-link-item';


describe('NodeDetailsHealthLinkItem', () => {
  describe('appendTime', () => {
    const time = '2017-06-01T00:00:00Z';
    const timeUnix = moment(time).unix();

    it('returns url for empty url or time', () => {
      expect(appendTime('', time)).toEqual('');
      expect(appendTime('foo', null)).toEqual('foo');
      expect(appendTime('', null)).toEqual('');
    });

    it('appends as json for cloud link', () => {
      const url = appendTime('/prom/:instanceid/notebook/new/%7B%22cells%22%3A%5B%7B%22queries%22%3A%5B%22go_goroutines%22%5D%7D%5D%7D', time);
      expect(url).toContain(timeUnix);

      const payload = JSON.parse(decodeURIComponent(url.substr(url.indexOf('new/') + 4)));
      expect(payload.time.queryEnd).toEqual(timeUnix);
    });

    it('appends as GET parameter', () => {
      expect(appendTime('http://example.test?q=foo', time)).toEqual('http://example.test?q=foo&time=1496275200');
      expect(appendTime('http://example.test/q=foo/', time)).toEqual('http://example.test/q=foo/?time=1496275200');
    });
  });
});
