import { fromJS, Map } from 'immutable';

const makeMap = Map;

describe('NodesLayout', () => {
  const NodesLayout = require('../nodes-layout');

  function getNodeCoordinates(nodes) {
    const coords = [];
    nodes
      .sortBy(node => node.get('id'))
      .forEach(node => {
        coords.push(node.get('x'));
        coords.push(node.get('y'));
      });
    return coords;
  }

  let options;
  let nodes;
  let coords;
  let resultCoords;

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
    options = {
      nodeCache: makeMap(),
      edgeCache: makeMap()
    };
  });

  it('detects unseen nodes', () => {
    const set1 = fromJS({
      n1: {id: 'n1'}
    });
    const set12 = fromJS({
      n1: {id: 'n1'},
      n2: {id: 'n2'}
    });
    const set13 = fromJS({
      n1: {id: 'n1'},
      n3: {id: 'n3'}
    });
    let hasUnseen;
    hasUnseen = NodesLayout.hasUnseenNodes(set12, set1);
    expect(hasUnseen).toBeTruthy();
    hasUnseen = NodesLayout.hasUnseenNodes(set13, set1);
    expect(hasUnseen).toBeTruthy();
    hasUnseen = NodesLayout.hasUnseenNodes(set1, set12);
    expect(hasUnseen).toBeFalsy();
    hasUnseen = NodesLayout.hasUnseenNodes(set1, set13);
    expect(hasUnseen).toBeFalsy();
    hasUnseen = NodesLayout.hasUnseenNodes(set12, set13);
    expect(hasUnseen).toBeTruthy();
  });

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

    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);
    coords = getNodeCoordinates(result.nodes);

    result = NodesLayout.doLayout(
      nodeSets.removeEdge24.nodes,
      nodeSets.removeEdge24.edges,
      options
    );
    nodes = result.nodes.toJS();
    // console.log('remove 1 edge', nodes, result);

    resultCoords = getNodeCoordinates(result.nodes);
    expect(resultCoords).toEqual(coords);
  });

  it('keeps nodes in rectangle after removed edge reappears', () => {
    let result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges);

    coords = getNodeCoordinates(result.nodes);
    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);
    result = NodesLayout.doLayout(
      nodeSets.removeEdge24.nodes,
      nodeSets.removeEdge24.edges,
      options
    );

    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);
    result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges,
      options
    );

    nodes = result.nodes.toJS();
    // console.log('re-add 1 edge', nodes, result);

    resultCoords = getNodeCoordinates(result.nodes);
    expect(resultCoords).toEqual(coords);
  });

  it('keeps nodes in rectangle after node disappears', () => {
    let result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges);

    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);
    result = NodesLayout.doLayout(
      nodeSets.removeNode2.nodes,
      nodeSets.removeNode2.edges,
      options
    );

    nodes = result.nodes.toJS();

    resultCoords = getNodeCoordinates(result.nodes);
    expect(resultCoords.slice(0, 2)).toEqual(coords.slice(0, 2));
    expect(resultCoords.slice(2, 6)).toEqual(coords.slice(4, 8));
  });

  it('keeps nodes in rectangle after removed node reappears', () => {
    let result = NodesLayout.doLayout(
      nodeSets.initial4.nodes,
      nodeSets.initial4.edges);

    nodes = result.nodes.toJS();

    coords = getNodeCoordinates(result.nodes);
    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);

    result = NodesLayout.doLayout(
      nodeSets.removeNode23.nodes,
      nodeSets.removeNode23.edges,
      options
    );

    nodes = result.nodes.toJS();

    expect(nodes.n1.x).toBeLessThan(nodes.n4.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n4.y);

    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);
    result = NodesLayout.doLayout(
      nodeSets.removeNode2.nodes,
      nodeSets.removeNode2.edges,
      options
    );

    nodes = result.nodes.toJS();
    // console.log('re-add 1 node', nodes);

    resultCoords = getNodeCoordinates(result.nodes);
    expect(resultCoords.slice(0, 2)).toEqual(coords.slice(0, 2));
    expect(resultCoords.slice(2, 6)).toEqual(coords.slice(4, 8));
  });

  it('renders single nodes in a square', () => {
    const result = NodesLayout.doLayout(
      nodeSets.single3.nodes,
      nodeSets.single3.edges);

    nodes = result.nodes.toJS();

    expect(nodes.n1.x).toEqual(nodes.n3.x);
    expect(nodes.n1.y).toEqual(nodes.n2.y);
    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n3.y);
  });

  it('renders single nodes next to portrait graph', () => {
    const result = NodesLayout.doLayout(
      nodeSets.singlePortrait.nodes,
      nodeSets.singlePortrait.edges);

    nodes = result.nodes.toJS();

    // first square row on same level as top-most other node
    expect(nodes.n1.y).toEqual(nodes.n2.y);
    expect(nodes.n1.y).toEqual(nodes.n3.y);
    expect(nodes.n4.y).toEqual(nodes.n5.y);

    // all singles right to other nodes
    expect(nodes.n1.x).toEqual(nodes.n4.x);
    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.x).toBeLessThan(nodes.n3.x);
    expect(nodes.n1.x).toBeLessThan(nodes.n5.x);
    expect(nodes.n2.x).toEqual(nodes.n5.x);
  });
});
