jest.dontMock('../node-details.js');
jest.dontMock('../../mixins/node-color-mixin');
jest.dontMock('../../utils/title-utils');

describe('NodeDetails', () => {
  let NodeDetails;
  let nodes;
  let nodeId;
  let details;
  const React = require('react');
  const Immutable = require('immutable');
  const TestUtils = require('react/lib/ReactTestUtils');
  const makeMap = Immutable.OrderedMap;

  beforeEach(() => {
    NodeDetails = require('../node-details.js');
    nodes = makeMap();
    nodeId = 'n1';
  });

  it('shows n/a when node was not found', () => {
    const c = TestUtils.renderIntoDocument(<NodeDetails nodes={nodes} nodeId={nodeId} />);
    const notFound = TestUtils.findRenderedDOMComponentWithClass(c, 'node-details-header-notavailable');
    expect(notFound).toBeDefined();
  });

  it('show label of node with title', () => {
    nodes = nodes.set(nodeId, Immutable.fromJS({id: nodeId}));
    details = {label_major: 'Node 1', tables: []};
    const c = TestUtils.renderIntoDocument(<NodeDetails nodes={nodes}
      nodeId={nodeId} details={details} />);

    const title = TestUtils.findRenderedDOMComponentWithClass(c, 'node-details-header-label');
    expect(title.textContent).toBe('Node 1');
  });
});
