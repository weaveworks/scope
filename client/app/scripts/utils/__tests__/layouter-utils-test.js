import { fromJS } from 'immutable';

describe('LayouterUtils', () => {
  const LayouterUtils = require('../layouter-utils');

  describe('initEdgesFromNodes', () => {
    const f = LayouterUtils.initEdgesFromNodes;

    it('should return map of edges', () => {
      const input = fromJS({
        a: { adjacency: ['b', 'c'] },
        b: { adjacency: ['a', 'b'] },
        c: {}
      });
      expect(f(input).toJS()).toEqual({
        'a-b': { id: 'a-b', source: 'a', target: 'b', value: 1 },
        'a-c': { id: 'a-c', source: 'a', target: 'c', value: 1 },
        'b-a': { id: 'b-a', source: 'b', target: 'a', value: 1 },
        'b-b': { id: 'b-b', source: 'b', target: 'b', value: 1 },
      });
    });
  });

  describe('collapseMultiEdges', () => {
    const f = LayouterUtils.collapseMultiEdges;

    it('should return collapsed multi-edges', () => {
      const input = fromJS({
        'a-b': { id: 'a-b', source: 'a', target: 'b' },
        'a-c': { id: 'a-c', source: 'a', target: 'c' },
        'b-a': { id: 'b-a', source: 'b', target: 'a' },
        'b-b': { id: 'b-b', source: 'b', target: 'b' },
      });
      expect(f(input).toJS()).toEqual({
        'a-b': { id: 'a-b', source: 'a', target: 'b', bidirectional: true },
        'a-c': { id: 'a-c', source: 'a', target: 'c' },
        'b-b': { id: 'b-b', source: 'b', target: 'b' },
      });
    });
  });
});
