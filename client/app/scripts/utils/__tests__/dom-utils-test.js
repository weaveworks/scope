import { encodeIdAttribute, decodeIdAttribute } from '../dom-utils';

describe('DomUtils', () => {
  describe('encodeIdAttribute/decodeIdAttribute', () => {
    it('encode should be reversible by decode ', () => {
      [
        '123-abc;<foo>',
        ';;<<><>',
        '!@#$%^&*()+-\'"',
      ].forEach((input) => {
        expect(decodeIdAttribute(encodeIdAttribute(input))).toEqual(input);
      });
    });
  });
});
