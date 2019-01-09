import { fromJS, Map } from 'immutable';

import { constructEdgeId as edge } from '../../utils/layouter-utils';

const makeMap = Map;

describe('NodesLayout', () => {
  const NodesLayout = require('../nodes-layout');

  function getNodeCoordinates(nodes) {
    const coords = [];
    nodes
      .sortBy(node => node.get('id'))
      .forEach((node) => {
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
      edges: fromJS({
        [edge('n1', 'n3')]: {id: edge('n1', 'n3'), source: 'n1', target: 'n3'},
        [edge('n1', 'n4')]: {id: edge('n1', 'n4'), source: 'n1', target: 'n4'},
        [edge('n2', 'n4')]: {id: edge('n2', 'n4'), source: 'n2', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      })
    },
    layoutProps: {
      edges: fromJS({}),
      nodes: fromJS({
        n1: {
          id: 'n1', label: 'lold', labelMinor: 'lmold', rank: 'rold'
        },
      })
    },
    layoutProps2: {
      edges: fromJS({}),
      nodes: fromJS({
        n1: {
          id: 'n1', label: 'lnew', labelMinor: 'lmnew', rank: 'rnew', x: 111, y: 109
        },
      })
    },
    rank4: {
      edges: fromJS({
        [edge('n1', 'n3')]: {id: edge('n1', 'n3'), source: 'n1', target: 'n3'},
        [edge('n1', 'n4')]: {id: edge('n1', 'n4'), source: 'n1', target: 'n4'},
        [edge('n2', 'n4')]: {id: edge('n2', 'n4'), source: 'n2', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1', rank: 'A'},
        n2: {id: 'n2', rank: 'A'},
        n3: {id: 'n3', rank: 'B'},
        n4: {id: 'n4', rank: 'B'}
      })
    },
    rank6: {
      edges: fromJS({
        [edge('n1', 'n3')]: {id: edge('n1', 'n3'), source: 'n1', target: 'n3'},
        [edge('n1', 'n4')]: {id: edge('n1', 'n4'), source: 'n1', target: 'n4'},
        [edge('n1', 'n5')]: {id: edge('n1', 'n5'), source: 'n1', target: 'n5'},
        [edge('n2', 'n4')]: {id: edge('n2', 'n4'), source: 'n2', target: 'n4'},
        [edge('n2', 'n6')]: {id: edge('n2', 'n6'), source: 'n2', target: 'n6'},
      }),
      nodes: fromJS({
        n1: {id: 'n1', rank: 'A'},
        n2: {id: 'n2', rank: 'A'},
        n3: {id: 'n3', rank: 'B'},
        n4: {id: 'n4', rank: 'B'},
        n5: {id: 'n5', rank: 'A'},
        n6: {id: 'n6', rank: 'B'},
      })
    },
    removeEdge24: {
      edges: fromJS({
        [edge('n1', 'n3')]: {id: edge('n1', 'n3'), source: 'n1', target: 'n3'},
        [edge('n1', 'n4')]: {id: edge('n1', 'n4'), source: 'n1', target: 'n4'}
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
        [edge('n1', 'n3')]: {id: edge('n1', 'n3'), source: 'n1', target: 'n3'},
        [edge('n1', 'n4')]: {id: edge('n1', 'n4'), source: 'n1', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      })
    },
    removeNode23: {
      edges: fromJS({
        [edge('n1', 'n4')]: {id: edge('n1', 'n4'), source: 'n1', target: 'n4'}
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
        [edge('n1', 'n4')]: {id: edge('n1', 'n4'), source: 'n1', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'},
        n5: {id: 'n5'}
      })
    },
    singlePortrait6: {
      edges: fromJS({
        [edge('n1', 'n4')]: {id: edge('n1', 'n4'), source: 'n1', target: 'n4'}
      }),
      nodes: fromJS({
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'},
        n5: {id: 'n5'},
        n6: {id: 'n6'}
      })
    }
  };

  beforeEach(() => {
    // clear feature flags
    window.localStorage.clear();

    options = {
      edgeCache: makeMap(),
      nodeCache: makeMap()
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
      nodeSets.initial4.edges
    );
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
      nodeSets.initial4.edges
    );

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
      nodeSets.initial4.edges
    );

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
      nodeSets.initial4.edges
    );

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
      nodeSets.initial4.edges
    );

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
      nodeSets.single3.edges
    );

    nodes = result.nodes.toJS();

    expect(nodes.n1.x).toEqual(nodes.n3.x);
    expect(nodes.n1.y).toEqual(nodes.n2.y);
    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n3.y);
  });

  it('renders single nodes next to portrait graph', () => {
    const result = NodesLayout.doLayout(
      nodeSets.singlePortrait.nodes,
      nodeSets.singlePortrait.edges,
      { noCache: true }
    );

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

  it('renders an additional single node in single nodes group', () => {
    let result = NodesLayout.doLayout(
      nodeSets.singlePortrait.nodes,
      nodeSets.singlePortrait.edges,
      { noCache: true }
    );

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

    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);

    result = NodesLayout.doLayout(
      nodeSets.singlePortrait6.nodes,
      nodeSets.singlePortrait6.edges,
      options
    );

    nodes = result.nodes.toJS();

    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.x).toBeLessThan(nodes.n3.x);
    expect(nodes.n1.x).toBeLessThan(nodes.n5.x);
    expect(nodes.n1.x).toBeLessThan(nodes.n6.x);
  });

  it('adds a new node to existing layout in a line', () => {
    // feature flag
    window.localStorage.setItem('scope-experimental:layout-dance', true);

    let result = NodesLayout.doLayout(
      nodeSets.rank4.nodes,
      nodeSets.rank4.edges,
      { noCache: true }
    );

    nodes = result.nodes.toJS();

    coords = getNodeCoordinates(result.nodes);
    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);

    expect(NodesLayout.hasNewNodesOfExistingRank(
      nodeSets.rank6.nodes,
      nodeSets.rank6.edges,
      result.nodes
    )).toBeTruthy();

    result = NodesLayout.doLayout(
      nodeSets.rank6.nodes,
      nodeSets.rank6.edges,
      options
    );

    nodes = result.nodes.toJS();

    expect(nodes.n5.x).toBeGreaterThan(nodes.n1.x);
    expect(nodes.n5.y).toEqual(nodes.n1.y);
    expect(nodes.n6.x).toBeGreaterThan(nodes.n3.x);
    expect(nodes.n6.y).toEqual(nodes.n3.y);
  });

  it('rerenders the nodes completely after the coordinates have been messed up', () => {
    // Take an initial setting
    let result = NodesLayout.doLayout(
      nodeSets.rank4.nodes,
      nodeSets.rank4.edges,
    );

    // Cache the result layout
    options.cachedLayout = result;
    options.nodeCache = options.nodeCache.merge(result.nodes);
    options.edgeCache = options.edgeCache.merge(result.edge);

    // Shrink the coordinates of all the notes 2x to make them closer to one another
    options.nodeCache = options.nodeCache.update(cache => cache.map(node => node.merge({
      x: node.get('x') / 2,
      y: node.get('y') / 2,
    })));

    // Rerun the initial layout to get a trivial diff and skip all the advanced layouting logic.
    result = NodesLayout.doLayout(
      nodeSets.rank4.nodes,
      nodeSets.rank4.edges,
      options
    );

    // The layout should have updated by running into our last 'integration testing' criterion
    coords = getNodeCoordinates(options.nodeCache);
    resultCoords = getNodeCoordinates(result.nodes);
    expect(resultCoords).not.toEqual(coords);
  });

  it('only caches layout-related properties', () => {
    // populate cache by doing a full layout run, this stores layout values in cache
    const first = NodesLayout.doLayout(
      nodeSets.layoutProps.nodes,
      nodeSets.layoutProps.edges,
      { noCache: true }
    );

    // now pass updated nodes with modified values
    const second = NodesLayout.doLayout(
      nodeSets.layoutProps2.nodes,
      nodeSets.layoutProps2.edges,
      {}
    );

    // new labels should not be overwritten by cache
    nodes = second.nodes.toJS();
    expect(nodes.n1.label).toEqual('lnew');
    expect(nodes.n1.labelMinor).toEqual('lmnew');
    // but layout values should be preferred from cache
    expect(nodes.n1.rank).toEqual('rold');
    expect(nodes.n1.x).toEqual(first.nodes.getIn(['n1', 'x']));
    expect(nodes.n1.y).toEqual(first.nodes.getIn(['n1', 'y']));
  });
});
