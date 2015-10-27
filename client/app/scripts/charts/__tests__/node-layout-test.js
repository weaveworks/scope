jest.dontMock('../nodes-layout');
jest.dontMock('../../constants/naming'); // edge naming: 'source-target'

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

  const nodeSets = {
    initial4: {
      nodes: {
        n1: {id: 'n1'},
        n2: {id: 'n2'},
        n3: {id: 'n3'},
        n4: {id: 'n4'}
      },
      edges: {
        'n1-n3': {id: 'n1-n3', source: {id: 'n1'}, target: {id: 'n3'}},
        'n1-n4': {id: 'n1-n4', source: {id: 'n1'}, target: {id: 'n4'}},
        'n2-n4': {id: 'n2-n4', source: {id: 'n2'}, target: {id: 'n4'}}
      }
    }
  };

  it('lays out initial nodeset', () => {
    const nodes = nodeSets.initial4.nodes;
    const edges = nodeSets.initial4.edges;
    NodesLayout.doLayout(nodes, edges);
    expect(nodes.n1.x).toBeLessThan(nodes.n2.x);
    expect(nodes.n1.y).toEqual(nodes.n2.y);

    expect(nodes.n1.x).toEqual(nodes.n3.x);
    expect(nodes.n1.y).toBeLessThan(nodes.n3.y);

    expect(nodes.n3.x).toBeLessThan(nodes.n4.x);
    expect(nodes.n3.y).toEqual(nodes.n4.y);

    console.log(nodes, nodeSets.initial4.nodes);
  });

});
