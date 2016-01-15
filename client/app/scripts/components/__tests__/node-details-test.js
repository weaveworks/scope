import React from 'react';
import Immutable from 'immutable';
import TestUtils from 'react/lib/ReactTestUtils';

jest.dontMock('../../dispatcher/app-dispatcher');
jest.dontMock('../node-details.js');
jest.dontMock('../node-details/node-details-controls.js');
jest.dontMock('../node-details/node-details-relatives.js');
jest.dontMock('../node-details/node-details-table.js');
jest.dontMock('../../utils/color-utils');
jest.dontMock('../../utils/title-utils');

// need ES5 require to keep automocking off
const NodeDetails = require('../node-details.js').default;

describe('NodeDetails', () => {
  let nodes;
  let nodeId;
  let details;
  const makeMap = Immutable.OrderedMap;

  beforeEach(() => {
    nodes = makeMap();
    nodeId = 'n1';
  });

  it('shows n/a when node was not found', () => {
    const c = TestUtils.renderIntoDocument(<NodeDetails notFound />);
    const notFound = TestUtils.findRenderedDOMComponentWithClass(c, 'node-details-header-notavailable');
    expect(notFound).toBeDefined();
  });

  it('show label of node with title', () => {
    nodes = nodes.set(nodeId, Immutable.fromJS({id: nodeId}));
    details = {label: 'Node 1'};
    const c = TestUtils.renderIntoDocument(<NodeDetails nodes={nodes}
      nodeId={nodeId} details={details} />);

    const title = TestUtils.findRenderedDOMComponentWithClass(c, 'node-details-header-label');
    expect(title.textContent).toBe('Node 1');
  });
});
