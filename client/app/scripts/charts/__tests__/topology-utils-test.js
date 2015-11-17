jest.dontMock('../topology-utils');
jest.dontMock('../../constants/naming'); // edge naming: 'source-target'

import { fromJS } from 'immutable';

describe('TopologyUtils', () => {
  let TopologyUtils;
  let nodes;

  const nodeSets = {
    initial4: {
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      }),
      edges: fromJS({
        'n1-n3': {id: 'n1-n3', source: 'n1', target: 'n3'},
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'},
        'n2-n4': {id: 'n2-n4', source: 'n2', target: 'n4'}
      })
    },
    removeEdge24: {
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      }),
      edges: fromJS({
        'n1-n3': {id: 'n1-n3', source: 'n1', target: 'n3'},
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      })
    },
    removeNode2: {
      nodes: fromJS({
        n1: {id: 'n1'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      }),
      edges: fromJS({
        'n1-n3': {id: 'n1-n3', source: 'n1', target: 'n3'},
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      })
    },
    removeNode23: {
      nodes: fromJS({
        n1: {id: 'n1'},
        n4: {id: 'n4'}
      }),
      edges: fromJS({
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      })
    },
    single3: {
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'}
      }),
      edges: fromJS({})
    },
    singlePortrait: {
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'},
        n5: {id: 'n5'}
      }),
      edges: fromJS({
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      })
    }
  };

  beforeEach(() => {
    TopologyUtils = require('../topology-utils');
  });

  it('sets node degrees', () => {
    nodes = TopologyUtils.updateNodeDegrees(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges).toJS();

    expect(nodes.n1.degree).toEqual(2);
    expect(nodes.n2.degree).toEqual(1);
    expect(nodes.n3.degree).toEqual(1);
    expect(nodes.n4.degree).toEqual(2);

    nodes = TopologyUtils.updateNodeDegrees(
      nodeSets.removeEdge24.nodes,
      nodeSets.removeEdge24.edges).toJS();

    expect(nodes.n1.degree).toEqual(2);
    expect(nodes.n2.degree).toEqual(0);
    expect(nodes.n3.degree).toEqual(1);
    expect(nodes.n4.degree).toEqual(1);

    nodes = TopologyUtils.updateNodeDegrees(
      nodeSets.single3.nodes,
      nodeSets.single3.edges).toJS();

    expect(nodes.n1.degree).toEqual(0);
    expect(nodes.n2.degree).toEqual(0);
    expect(nodes.n3.degree).toEqual(0);
  });
});
