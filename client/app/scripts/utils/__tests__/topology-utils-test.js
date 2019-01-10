import { fromJS } from 'immutable';

describe('TopologyUtils', () => {
  let TopologyUtils;
  let nodes;

  const nodeSets = {
    initial4: {
      edges: fromJS({
        'n1-n3': {id: 'n1-n3', source: 'n1', target: 'n3'},
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'},
        'n2-n4': {id: 'n2-n4', source: 'n2', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      })
    },
    removeEdge24: {
      edges: fromJS({
        'n1-n3': {id: 'n1-n3', source: 'n1', target: 'n3'},
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      })
    },
    removeNode2: {
      edges: fromJS({
        'n1-n3': {id: 'n1-n3', source: 'n1', target: 'n3'},
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      })
    },
    removeNode23: {
      edges: fromJS({
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n4: {id: 'n4'}
      })
    },
    single3: {
      edges: fromJS({}),
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'}
      })
    },
    singlePortrait: {
      edges: fromJS({
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'},
        n5: {id: 'n5'}
      })
    }
  };

  beforeEach(() => {
    TopologyUtils = require('../topology-utils');
  });

  it('sets node degrees', () => {
    nodes = TopologyUtils.updateNodeDegrees(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges
    ).toJS();

    expect(nodes.n1.degree).toEqual(2);
    expect(nodes.n2.degree).toEqual(1);
    expect(nodes.n3.degree).toEqual(1);
    expect(nodes.n4.degree).toEqual(2);

    nodes = TopologyUtils.updateNodeDegrees(
      nodeSets.removeEdge24.nodes,
      nodeSets.removeEdge24.edges
    ).toJS();

    expect(nodes.n1.degree).toEqual(2);
    expect(nodes.n2.degree).toEqual(0);
    expect(nodes.n3.degree).toEqual(1);
    expect(nodes.n4.degree).toEqual(1);

    nodes = TopologyUtils.updateNodeDegrees(
      nodeSets.single3.nodes,
      nodeSets.single3.edges
    ).toJS();

    expect(nodes.n1.degree).toEqual(0);
    expect(nodes.n2.degree).toEqual(0);
    expect(nodes.n3.degree).toEqual(0);
  });

  describe('buildTopologyCacheId', () => {
    it('should generate a cache ID', () => {
      const fun = TopologyUtils.buildTopologyCacheId;
      expect(fun()).toEqual('');
      expect(fun('test')).toEqual('test');
      expect(fun(undefined, 'test')).toEqual('');
      expect(fun('test', {a: 1})).toEqual('test{"a":1}');
    });
  });

  describe('filterHiddenTopologies', () => {
    it('should filter out empty topos that set hide_if_empty=true', () => {
      const topos = [
        {hide_if_empty: true, id: 'a', stats: {filtered_nodes: 0, node_count: 0}},
        {hide_if_empty: true, id: 'b', stats: {filtered_nodes: 0, node_count: 1}},
        {hide_if_empty: true, id: 'c', stats: {filtered_nodes: 1, node_count: 0}},
        {hide_if_empty: false, id: 'd', stats: {filtered_nodes: 0, node_count: 0}}
      ];

      const res = TopologyUtils.filterHiddenTopologies(topos);
      expect(res.map(t => t.id)).toEqual(['b', 'c', 'd']);
    });
  });
});
