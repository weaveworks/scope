import { fromJS } from 'immutable';

import {
  initEdgesFromNodes,
} from '../layouter-utils';


describe('LayouterUtils', () => {
  describe('initEdgesFromNodes', () => {
    it('should return map of edges', () => {
      const input = fromJS({
        a: { adjacency: ['b', 'c'] },
        b: { adjacency: ['a', 'b'] },
        c: {}
      });
      expect(initEdgesFromNodes(input).toJS()).toEqual({
        'a-b': { id: 'a-b', source: 'a', target: 'b', value: 1 },
        'a-c': { id: 'a-c', source: 'a', target: 'c', value: 1 },
        'b-a': { id: 'b-a', source: 'b', target: 'a', value: 1 },
        'b-b': { id: 'b-b', source: 'b', target: 'b', value: 1 },
      });
    });
  });
});
