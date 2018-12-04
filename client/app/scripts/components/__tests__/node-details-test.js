import React from 'react';
import Immutable from 'immutable';
import TestUtils from 'react-dom/test-utils';
import { Provider } from 'react-redux';
import configureStore from '../../stores/configureStore';

// need ES5 require to keep automocking off
const NodeDetails = require('../node-details.js').default.WrappedComponent;

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
    const c = TestUtils.renderIntoDocument((
      <Provider store={configureStore()}>
        <NodeDetails notFound />
      </Provider>
    ));
    const notFound = TestUtils.findRenderedDOMComponentWithClass(
      c,
      'node-details-header-notavailable'
    );
    expect(notFound).toBeDefined();
  });

  it('show label of node with title', () => {
    nodes = nodes.set(nodeId, Immutable.fromJS({id: nodeId}));
    details = {label: 'Node 1'};
    const c = TestUtils.renderIntoDocument((
      <Provider store={configureStore()}>
        <NodeDetails
          nodes={nodes}
          topologyId="containers"
          nodeId={nodeId}
          details={details}
          />
      </Provider>
    ));

    const title = TestUtils.findRenderedDOMComponentWithClass(c, 'node-details-header-label');
    expect(title.title).toBe('Node 1');
  });
});
