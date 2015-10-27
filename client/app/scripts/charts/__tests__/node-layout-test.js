jest.dontMock('../nodes-layout');
jest.dontMock('../../constants/naming'); // edge naming: 'source-target'

import { fromJS } from 'immutable';

describe('NodesLayout', () => {
  const NodesLayout = require('../nodes-layout');

  function scale(val) {
    return val * 3;
  }
  const topologyId = 'tid';
  const width = 80;
  const height = 80;
  const margins = {
    left: 0,
    top: 0
  };
  let history;
  let nodes;

  const nodeSets = {
    initial4: {
      nodes: {
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      },
      edges: {
        'n1-n3': {id: 'n1-n3', source: 'n1', target: 'n3'},
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'},
        'n2-n4': {id: 'n2-n4', source: 'n2', target: 'n4'}
      }
    },
    removeEdge24: {
      nodes: {
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      },
      edges: {
        'n1-n3': {id: 'n1-n3', source: 'n1', target: 'n3'},
        'n1-n4': {id: 'n1-n4', source: 'n1', target: 'n4'}
      }
    }
  };

  it('lays out initial nodeset in a rectangle', () => {
    const result = NodesLayout.doLayout(
      fromJS(nodeSets.initial4.nodes),
      fromJS(nodeSets.initial4.edges));
    // console.log('initial', result.get('nodes'));
    nodes = result.nodes.toJS();

    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.y).toEqual(nodes.n2.y);
    expect(nodes.n1.x).toEqual(nodes.n3.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n3.y);
    expect(nodes.n3.x).toBeLessThan(nodes.n4.x);
    expect(nodes.n3.y).toEqual(nodes.n4.y);
  });

  // it('keeps nodes in rectangle after removing one edge', () => {
  //   history = [{
  //     nodes: nodeSets.initial4.nodes,
  //     edges: nodeSets.initial4.edges
  //   }];
  //   nodes = nodeSets.removeEdge24.nodes;
  //   edges = nodeSets.removeEdge24.edges;
  //   NodesLayout.doLayout(nodes, edges, {history});
  //   console.log('remove 1 edge', nodes);
  //
  //   expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
  //   expect(nodes.n1.y).toEqual(nodes.n2.y);
  //   expect(nodes.n1.x).toEqual(nodes.n3.x);
  //   expect(nodes.n1.y).toBeLessThan(nodes.n3.y);
  //   expect(nodes.n3.x).toBeLessThan(nodes.n4.x);
  //   expect(nodes.n3.y).toEqual(nodes.n4.y);
  //
  // });

});
