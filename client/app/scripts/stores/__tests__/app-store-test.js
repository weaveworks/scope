jest.dontMock('../../utils/topology-utils');
jest.dontMock('../../constants/action-types');
jest.dontMock('../app-store');

// Appstore test suite using Jasmine matchers

describe('AppStore', () => {
  const ActionTypes = require('../../constants/action-types').default;
  let AppStore;
  let registeredCallback;

  // fixtures

  const NODE_SET = {
    n1: {
      id: 'n1',
      rank: undefined,
      adjacency: ['n1', 'n2'],
      pseudo: undefined,
      label: undefined,
      label_minor: undefined
    },
    n2: {
      id: 'n2',
      rank: undefined,
      adjacency: undefined,
      pseudo: undefined,
      label: undefined,
      label_minor: undefined
    }
  };

  // actions

  const ChangeTopologyOptionAction = {
    type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
    topologyId: 'topo1',
    option: 'option1',
    value: 'on'
  };

  const ChangeTopologyOptionAction2 = {
    type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
    topologyId: 'topo1',
    option: 'option1',
    value: 'off'
  };

  const ClickNodeAction = {
    type: ActionTypes.CLICK_NODE,
    nodeId: 'n1'
  };

  const ClickNode2Action = {
    type: ActionTypes.CLICK_NODE,
    nodeId: 'n2'
  };

  const ClickRelativeAction = {
    type: ActionTypes.CLICK_RELATIVE,
    nodeId: 'rel1'
  };

  const ClickShowTopologyForNodeAction = {
    type: ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE,
    topologyId: 'topo2',
    nodeId: 'rel1'
  };

  const ClickSubTopologyAction = {
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId: 'topo1-grouped'
  };

  const ClickTopologyAction = {
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId: 'topo1'
  };

  const ClickTopology2Action = {
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId: 'topo2'
  };

  const CloseWebsocketAction = {
    type: ActionTypes.CLOSE_WEBSOCKET
  };

  const deSelectNode = {
    type: ActionTypes.DESELECT_NODE
  };

  const OpenWebsocketAction = {
    type: ActionTypes.OPEN_WEBSOCKET
  };

  const ReceiveNodesDeltaAction = {
    type: ActionTypes.RECEIVE_NODES_DELTA,
    delta: {
      add: [{
        id: 'n1',
        adjacency: ['n1', 'n2']
      }, {
        id: 'n2'
      }]
    }
  };

  const ReceiveNodesDeltaUpdateAction = {
    type: ActionTypes.RECEIVE_NODES_DELTA,
    delta: {
      update: [{
        id: 'n1',
        adjacency: ['n1']
      }],
      remove: ['n2']
    }
  };

  const ReceiveTopologiesAction = {
    type: ActionTypes.RECEIVE_TOPOLOGIES,
    topologies: [{
      url: '/topo1',
      name: 'Topo1',
      options: {
        option1: [
          {value: 'on'},
          {value: 'off', default: true}
        ]
      },
      stats: {
        node_count: 1
      },
      sub_topologies: [{
        url: '/topo1-grouped',
        name: 'topo 1 grouped'
      }]
    }, {
      url: '/topo2',
      name: 'Topo2',
      stats: {
        node_count: 0
      }
    }]
  };

  const RouteAction = {
    type: ActionTypes.ROUTE_TOPOLOGY,
    state: {}
  };

  beforeEach(() => {
    AppStore = require('../app-store').default;
    const AppDispatcher = AppStore.getDispatcher();
    const callback = AppDispatcher.dispatch.bind(AppDispatcher);
    registeredCallback = callback;
  });

  // topology tests

  it('init with no topologies', () => {
    const topos = AppStore.getTopologies();
    expect(topos.size).toBe(0);
    expect(AppStore.getCurrentTopology()).toBeUndefined();
  });

  it('get current topology', () => {
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveTopologiesAction);

    expect(AppStore.getTopologies().size).toBe(2);
    expect(AppStore.getCurrentTopology().get('name')).toBe('Topo1');
    expect(AppStore.getCurrentTopologyUrl()).toBe('/topo1');
    expect(AppStore.getCurrentTopologyOptions().get('option1')).toBeDefined();
  });

  it('get sub-topology', () => {
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickSubTopologyAction);

    expect(AppStore.getTopologies().size).toBe(2);
    expect(AppStore.getCurrentTopology().get('name')).toBe('topo 1 grouped');
    expect(AppStore.getCurrentTopologyUrl()).toBe('/topo1-grouped');
    expect(AppStore.getCurrentTopologyOptions().size).toBe(0);
  });

  // topology options

  it('changes topology option', () => {
    // default options
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickTopologyAction);
    expect(AppStore.getActiveTopologyOptions().has('option1')).toBeTruthy();
    expect(AppStore.getActiveTopologyOptions().get('option1')).toBe('off');
    expect(AppStore.getAppState().topologyOptions.topo1.option1).toBe('off');

    // turn on
    registeredCallback(ChangeTopologyOptionAction);
    expect(AppStore.getActiveTopologyOptions().get('option1')).toBe('on');
    expect(AppStore.getAppState().topologyOptions.topo1.option1).toBe('on');

    // turn off
    registeredCallback(ChangeTopologyOptionAction2);
    expect(AppStore.getActiveTopologyOptions().get('option1')).toBe('off');
    expect(AppStore.getAppState().topologyOptions.topo1.option1).toBe('off');

    // other topology w/o options dont return options, but keep in app state
    registeredCallback(ClickSubTopologyAction);
    expect(AppStore.getActiveTopologyOptions()).toBeUndefined();
    expect(AppStore.getAppState().topologyOptions.topo1.option1).toBe('off');
  });

  it('sets topology options from route', () => {
    RouteAction.state = {
      topologyId: 'topo1',
      selectedNodeId: null,
      topologyOptions: {topo1: {option1: 'on'}}};
    registeredCallback(RouteAction);
    expect(AppStore.getActiveTopologyOptions().get('option1')).toBe('on');
    expect(AppStore.getAppState().topologyOptions.topo1.option1).toBe('on');

    // stay same after topos have been received
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickTopologyAction);
    expect(AppStore.getActiveTopologyOptions().get('option1')).toBe('on');
    expect(AppStore.getAppState().topologyOptions.topo1.option1).toBe('on');
  });

  it('uses default topology options from route', () => {
    RouteAction.state = {
      topologyId: 'topo1',
      selectedNodeId: null,
      topologyOptions: null};
    registeredCallback(RouteAction);
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickTopologyAction);
    expect(AppStore.getActiveTopologyOptions().get('option1')).toBe('off');
    expect(AppStore.getAppState().topologyOptions.topo1.option1).toBe('off');
  });

  // nodes delta

  it('replaces adjacency on update', () => {
    registeredCallback(ReceiveNodesDeltaAction);
    expect(AppStore.getNodes().toJS().n1.adjacency).toEqual(['n1', 'n2']);
    registeredCallback(ReceiveNodesDeltaUpdateAction);
    expect(AppStore.getNodes().toJS().n1.adjacency).toEqual(['n1']);
  });

  // browsing

  it('shows nodes that were received', () => {
    registeredCallback(ReceiveNodesDeltaAction);
    expect(AppStore.getNodes().toJS()).toEqual(NODE_SET);
  });

  it('knows a route was set', () => {
    expect(AppStore.isRouteSet()).toBeFalsy();
    registeredCallback(RouteAction);
    expect(AppStore.isRouteSet()).toBeTruthy();
  });

  it('gets selected node after click', () => {
    registeredCallback(ReceiveNodesDeltaAction);

    registeredCallback(ClickNodeAction);
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodes().toJS()).toEqual(NODE_SET);

    registeredCallback(deSelectNode);
    expect(AppStore.getSelectedNodeId()).toBe(null);
    expect(AppStore.getNodes().toJS()).toEqual(NODE_SET);
  });

  it('keeps showing nodes on navigating back after node click', () => {
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveNodesDeltaAction);

    expect(AppStore.getAppState().selectedNodeId).toEqual(null);

    registeredCallback(ClickNodeAction);
    expect(AppStore.getAppState().selectedNodeId).toEqual('n1');

    // go back in browsing
    RouteAction.state = {topologyId: 'topo1', selectedNodeId: null};
    registeredCallback(RouteAction);
    expect(AppStore.getSelectedNodeId()).toBe(null);
    expect(AppStore.getNodes().toJS()).toEqual(NODE_SET);
  });

  it('closes details when changing topologies', () => {
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveNodesDeltaAction);

    expect(AppStore.getAppState().selectedNodeId).toEqual(null);
    expect(AppStore.getAppState().topologyId).toEqual('topo1');

    registeredCallback(ClickNodeAction);
    expect(AppStore.getAppState().selectedNodeId).toEqual('n1');
    expect(AppStore.getAppState().topologyId).toEqual('topo1');

    registeredCallback(ClickSubTopologyAction);
    expect(AppStore.getAppState().selectedNodeId).toEqual(null);
    expect(AppStore.getAppState().topologyId).toEqual('topo1-grouped');
  });

  // connection errors

  it('resets topology on websocket reconnect', () => {
    registeredCallback(ReceiveNodesDeltaAction);
    expect(AppStore.getNodes().toJS()).toEqual(NODE_SET);

    registeredCallback(CloseWebsocketAction);
    expect(AppStore.isWebsocketClosed()).toBeTruthy();
    // keep showing old nodes
    expect(AppStore.getNodes().toJS()).toEqual(NODE_SET);

    registeredCallback(OpenWebsocketAction);
    expect(AppStore.isWebsocketClosed()).toBeFalsy();
    // opened socket clears nodes
    expect(AppStore.getNodes().toJS()).toEqual({});
  });

  // adjacency test

  it('returns the correct adjacency set for a node', () => {
    registeredCallback(ReceiveNodesDeltaAction);
    expect(AppStore.getAdjacentNodes().size).toEqual(0);

    registeredCallback(ClickNodeAction);
    expect(AppStore.getAdjacentNodes('n1').size).toEqual(2);
    expect(AppStore.getAdjacentNodes('n1').has('n1')).toBeTruthy();
    expect(AppStore.getAdjacentNodes('n1').has('n2')).toBeTruthy();

    registeredCallback(deSelectNode);
    expect(AppStore.getAdjacentNodes().size).toEqual(0);
  });

  // empty topology

  it('detects that the topology is empty', () => {
    registeredCallback(ReceiveTopologiesAction);
    registeredCallback(ClickTopologyAction);
    expect(AppStore.isTopologyEmpty()).toBeFalsy();

    registeredCallback(ClickTopology2Action);
    expect(AppStore.isTopologyEmpty()).toBeTruthy();

    registeredCallback(ClickTopologyAction);
    expect(AppStore.isTopologyEmpty()).toBeFalsy();
  });

  // selection of relatives

  it('keeps relatives as a stack', () => {
    registeredCallback(ClickNodeAction);
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodeDetails().size).toEqual(1);
    expect(AppStore.getNodeDetails().has('n1')).toBeTruthy();
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('n1');

    registeredCallback(ClickRelativeAction);
    // stack relative, first node stays main node
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('rel1');
    expect(AppStore.getNodeDetails().size).toEqual(2);
    expect(AppStore.getNodeDetails().has('rel1')).toBeTruthy();

    // click on first node should clear the stack
    registeredCallback(ClickNodeAction);
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('n1');
    expect(AppStore.getNodeDetails().size).toEqual(1);
    expect(AppStore.getNodeDetails().has('rel1')).toBeFalsy();
  });

  it('keeps clears stack when sibling is clicked', () => {
    registeredCallback(ClickNodeAction);
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodeDetails().size).toEqual(1);
    expect(AppStore.getNodeDetails().has('n1')).toBeTruthy();
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('n1');

    registeredCallback(ClickRelativeAction);
    // stack relative, first node stays main node
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('rel1');
    expect(AppStore.getNodeDetails().size).toEqual(2);
    expect(AppStore.getNodeDetails().has('rel1')).toBeTruthy();

    // click on sibling node should clear the stack
    registeredCallback(ClickNode2Action);
    expect(AppStore.getSelectedNodeId()).toBe('n2');
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('n2');
    expect(AppStore.getNodeDetails().size).toEqual(1);
    expect(AppStore.getNodeDetails().has('n1')).toBeFalsy();
    expect(AppStore.getNodeDetails().has('rel1')).toBeFalsy();
  });

  it('selectes relatives topology while keeping node selected', () => {
    registeredCallback(ClickTopologyAction);
    registeredCallback(ReceiveTopologiesAction);
    expect(AppStore.getCurrentTopology().get('name')).toBe('Topo1');

    registeredCallback(ClickNodeAction);
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodeDetails().size).toEqual(1);
    expect(AppStore.getNodeDetails().has('n1')).toBeTruthy();
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('n1');

    registeredCallback(ClickRelativeAction);
    // stack relative, first node stays main node
    expect(AppStore.getSelectedNodeId()).toBe('n1');
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('rel1');
    expect(AppStore.getNodeDetails().size).toEqual(2);
    expect(AppStore.getNodeDetails().has('rel1')).toBeTruthy();

    // click switches over to relative's topology and selectes relative
    registeredCallback(ClickShowTopologyForNodeAction);
    expect(AppStore.getSelectedNodeId()).toBe('rel1');
    expect(AppStore.getNodeDetails().keySeq().last()).toEqual('rel1');
    expect(AppStore.getNodeDetails().size).toEqual(1);
    expect(AppStore.getCurrentTopology().get('name')).toBe('Topo2');
  });
});
