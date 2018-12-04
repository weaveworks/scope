import { fromJS } from 'immutable';

import {
  initEdgesFromNodes,
  constructEdgeId as edge
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
        [edge('a', 'b')]: {
          id: edge('a', 'b'), source: 'a', target: 'b', value: 1
        },
        [edge('a', 'c')]: {
          id: edge('a', 'c'), source: 'a', target: 'c', value: 1
        },
        [edge('b', 'a')]: {
          id: edge('b', 'a'), source: 'b', target: 'a', value: 1
        },
        [edge('b', 'b')]: {
          id: edge('b', 'b'), source: 'b', target: 'b', value: 1
        },
      });
    });
  });
});
