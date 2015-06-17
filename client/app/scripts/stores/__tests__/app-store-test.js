

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

  const ClickSubTopologyAction = {
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId: 'topo1-grouped'
  };

  const ClickTopologyAction = {
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId: 'topo1'
  };

  const ClickGroupingAction = {
    type: ActionTypes.CLICK_GROUPING,
    grouping: 'grouped'
  };

  const CloseWebsocketAction = {
    type: ActionTypes.CLOSE_WEBSOCKET
  };

  const HitEscAction = {
    type: ActionTypes.HIT_ESC_KEY
  };

  const ReceiveEmptyNodesDeltaAction = {
    type: ActionTypes.RECEIVE_NODES_DELTA,
    delta: {}
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
      name: 'Topo1',
      sub_topologies: [{
        url: '/topo1-grouped',
        name: 'topo 1 grouped'
      }]
    }]
  };

  const RouteAction = {
    type: ActionTypes.ROUTE_TOPOLOGY,
    state: {}
  };

  beforeEach(function() {
    // clear AppStore singleton
    delete require.cache[require.resolve('../app-store')];
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

  it('get sub-topology', function() {
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickSubTopologyAction);

    expect(AppStore.getTopologies().length).toBe(1);
    expect(AppStore.getCurrentTopology().name).toBe('topo 1 grouped');
    expect(AppStore.getCurrentTopologyUrl()).toBe('/topo1-grouped');
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
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveNodesDeltaAction);

    expect(AppStore.getAppState())
      .toEqual({"topologyId":"topo1","selectedNodeId": null});

    registeredCallback(ClickNodeAction);
    expect(AppStore.getAppState())
      .toEqual({"topologyId":"topo1","selectedNodeId": 'n1'});

    // go back in browsing
    RouteAction.state = {"topologyId":"topo1","selectedNodeId": null};
    registeredCallback(RouteAction);
    expect(AppStore.getSelectedNodeId()).toBe(null);
    expect(AppStore.getNodes()).toEqual(NODE_SET);
  });

  // connection errors

  it('resets topology on websocket reconnect', function() {
    registeredCallback(ReceiveNodesDeltaAction);
    expect(AppStore.getNodes()).toEqual(NODE_SET);

    registeredCallback(CloseWebsocketAction);
    expect(AppStore.isWebsocketClosed()).toBeTruthy();
    expect(AppStore.getNodes()).toEqual(NODE_SET);

    registeredCallback(ReceiveEmptyNodesDeltaAction);
    expect(AppStore.getNodes()).toEqual({});
  });


});