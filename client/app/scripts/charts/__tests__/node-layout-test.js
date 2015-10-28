jest.dontMock('../nodes-layout');
jest.dontMock('../../constants/naming'); // edge naming: 'source-target'

import { fromJS, List } from 'immutable';

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
  let history = List();
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
  };

  beforeEach(() => {
    history = history.clear();
  })

  it('lays out initial nodeset in a rectangle', () => {
    const result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges);
    // console.log('initial', result.get('nodes'));
    nodes = result.nodes.toJS();

    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.y).toEqual(nodes.n2.y);
    expect(nodes.n1.x).toEqual(nodes.n3.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n3.y);
    expect(nodes.n3.x).toBeLessThan(nodes.n4.x);
    expect(nodes.n3.y).toEqual(nodes.n4.y);
  });

  it('keeps nodes in rectangle after removing one edge', () => {
    let result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges);
    history = history.unshift({
      nodes: result.nodes,
      edges: result.edges
    });
    result = NodesLayout.doLayout(
      nodeSets.removeEdge24.nodes,
      nodeSets.removeEdge24.edges,
      {history}
    );
    nodes = result.nodes.toJS();
    // console.log('remove 1 edge', nodes, result);

    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.y).toEqual(nodes.n2.y);
    expect(nodes.n1.x).toEqual(nodes.n3.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n3.y);
    expect(nodes.n3.x).toBeLessThan(nodes.n4.x);
    expect(nodes.n3.y).toEqual(nodes.n4.y);
  });

  it('keeps nodes in rectangle after removed edge reappears', () => {
    let result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges);

    history = history.unshift({
      nodes: result.nodes,
      edges: result.edges
    });
    result = NodesLayout.doLayout(
      nodeSets.removeEdge24.nodes,
      nodeSets.removeEdge24.edges,
      {history}
    );

    history = history.unshift({
      nodes: result.nodes,
      edges: result.edges
    });
    result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges,
      {history}
    );

    nodes = result.nodes.toJS();
    // console.log('re-add 1 edge', nodes, result);

    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.y).toEqual(nodes.n2.y);
    expect(nodes.n1.x).toEqual(nodes.n3.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n3.y);
    expect(nodes.n3.x).toBeLessThan(nodes.n4.x);
    expect(nodes.n3.y).toEqual(nodes.n4.y);
  });

  it('keeps nodes in rectangle after node dissappears', () => {
    let result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges);

    history = history.unshift({
      nodes: result.nodes,
      edges: result.edges
    });
    result = NodesLayout.doLayout(
      nodeSets.removeNode2.nodes,
      nodeSets.removeNode2.edges,
      {history}
    );

    nodes = result.nodes.toJS();

    expect(nodes.n1.x).toEqual(nodes.n3.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n3.y);
    expect(nodes.n3.x).toBeLessThan(nodes.n4.x);
    expect(nodes.n3.y).toEqual(nodes.n4.y);
  });

});
