

describe('AppStore', function() {
  const ActionTypes = require('../../constants/action-types');
  let AppStore;
  let registeredCallback;

  // fixtures

  const NODE_SET = {n1: {id: 'n1'}, n2: {id: 'n2'}};

  // actions

  const ClickNodeAction = {
    type: ActionTypes.CLICK_NODE,
    nodeId: 'n1'
  };

  const ClickTopologyAction = {
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId: 'topo1'
  };

  const ClickGroupingAction = {
    type: ActionTypes.CLICK_GROUPING,
    grouping: 'grouped'
  };

  const HitEscAction = {
    type: ActionTypes.HIT_ESC_KEY
  };

  const ReceiveNodesDeltaAction = {
    type: ActionTypes.RECEIVE_NODES_DELTA,
    delta: {
      add: [{
        id: 'n1'
      }, {
        id: 'n2'
      }]
    }
  };

  const ReceiveTopologiesAction = {
    type: ActionTypes.RECEIVE_TOPOLOGIES,
    topologies: [{
      url: '/topo1',
      grouped_url: '/topo1grouped',
      name: 'Topo1'
    }]
  };

  const RouteAction = {
    type: ActionTypes.ROUTE_TOPOLOGY,
    state: {}
  };

  beforeEach(function() {
    AppStore = require('../app-store');
    registeredCallback = AppStore.registeredCallback;
  });

  // topology tests

  it('init with no topologies', function() {
    const topos = AppStore.getTopologies();
    expect(topos.length).toBe(0);
    expect(AppStore.getCurrentTopology()).toBeUndefined();
  });

  it('get current topology', function() {
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveTopologiesAction);

    expect(AppStore.getTopologies().length).toBe(1);
    expect(AppStore.getCurrentTopology().name).toBe('Topo1');
    expect(AppStore.getCurrentTopologyUrl()).toBe('/topo1');
  });

  it('get grouped topology', function() {
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickGroupingAction);

    expect(AppStore.getTopologies().length).toBe(1);
    expect(AppStore.getCurrentTopology().name).toBe('Topo1');
    expect(AppStore.getCurrentTopologyUrl()).toBe('/topo1grouped');
  });

  // browsing

  it('shows nodes that were received', function() {
    registeredCallback(ReceiveNodesDeltaAction);
    expect(AppStore.getNodes()).toEqual(NODE_SET);
  });

  it('gets selected node after click', function() {
    registeredCallback(ReceiveNodesDeltaAction);

    registeredCallback(ClickNodeAction);
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodes()).toEqual(NODE_SET);

    registeredCallback(HitEscAction)
    expect(AppStore.getSelectedNodeId()).toBe(null);
    expect(AppStore.getNodes()).toEqual(NODE_SET);
  });

  it('keeps showing nodes on navigating back after node click', function() {
    registeredCallback(ReceiveNodesDeltaAction);
    // TODO clear AppStore cache
    expect(AppStore.getAppState())
      .toEqual({"topologyId":"topo1","grouping":"grouped","selectedNodeId": null});

    registeredCallback(ClickNodeAction);
    expect(AppStore.getAppState())
      .toEqual({"topologyId":"topo1","grouping":"grouped","selectedNodeId": 'n1'});

    // go back in browsing
    RouteAction.state = {"topologyId":"topo1","grouping":"grouped","selectedNodeId": null};
    registeredCallback(RouteAction);
    expect(AppStore.getSelectedNodeId()).toBe(null);
    expect(AppStore.getNodes()).toEqual(NODE_SET);

  });


});