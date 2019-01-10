import { is, fromJS } from 'immutable';

import { TABLE_VIEW_MODE } from '../../constants/naming';
import { constructEdgeId } from '../../utils/layouter-utils';
import { highlightedEdgeIdsSelector } from '../../selectors/graph-view/decorators';


// Root reducer test suite using Jasmine matchers
describe('RootReducer', () => {
  const ActionTypes = require('../../constants/action-types').default;
  const reducer = require('../root').default;
  const { initialState } = require('../root');
  const topologyUtils = require('../../utils/topology-utils');
  const topologySelectors = require('../../selectors/topology');
  // TODO maybe extract those to topology-utils tests?
  const { activeTopologyOptionsSelector } = topologySelectors;
  const { getAdjacentNodes, isNodesDisplayEmpty, isTopologyNodeCountZero } = topologyUtils;
  const { getUrlState } = require('../../utils/router-utils');

  // fixtures

  const NODE_SET = {
    n1: {
      adjacency: ['n1', 'n2'],
      filtered: false,
      id: 'n1',
    },
    n2: {
      filtered: false,
      id: 'n2',
    }
  };

  const topologies = [
    {
      fullName: 'Processes',
      hide_if_empty: true,
      id: 'processes',
      name: 'Processes',
      options: [
        {
          defaultValue: 'hide',
          id: 'unconnected',
          options: [
            {
              label: 'Unconnected nodes hidden',
              value: 'hide'
            }
          ],
          selectType: 'one'
        }
      ],
      rank: 1,
      stats: {
        edge_count: 379,
        filtered_nodes: 214,
        node_count: 320,
        nonpseudo_node_count: 320
      },
      sub_topologies: [],
      url: '/api/topology/processes'
    },
    {
      hide_if_empty: true,
      name: 'Pods',
      options: [
        {
          defaultValue: 'default',
          id: 'namespace',
          options: [
            {
              label: 'monitoring',
              value: 'monitoring'
            },
            {
              label: 'scope',
              value: 'scope'
            },
            {
              label: 'All Namespaces',
              value: 'all'
            }
          ],
          selectType: 'many'
        },
        {
          defaultValue: 'hide',
          id: 'pseudo',
          options: [
            {
              label: 'Show Unmanaged',
              value: 'show'
            },
            {
              label: 'Hide Unmanaged',
              value: 'hide'
            }
          ]
        }
      ],
      rank: 3,
      stats: {
        edge_count: 15,
        filtered_nodes: 16,
        node_count: 32,
        nonpseudo_node_count: 27
      },
      sub_topologies: [
        {
          hide_if_empty: true,
          name: 'services',
          options: [
            {
              defaultValue: 'default',
              id: 'namespace',
              options: [
                {
                  label: 'monitoring',
                  value: 'monitoring'
                },
                {
                  label: 'scope',
                  value: 'scope'
                },
                {
                  label: 'All Namespaces',
                  value: 'all'
                }
              ],
              selectType: 'many'
            }
          ],
          rank: 0,
          stats: {
            edge_count: 14,
            filtered_nodes: 16,
            node_count: 159,
            nonpseudo_node_count: 154
          },
          url: '/api/topology/services'
        }
      ],
      url: '/api/topology/pods'
    }
  ];

  // actions

  const ChangeTopologyOptionAction = {
    option: 'option1',
    topologyId: 'topo1',
    type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
    value: ['on']
  };

  const ChangeTopologyOptionAction2 = {
    option: 'option1',
    topologyId: 'topo1',
    type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
    value: ['off']
  };

  const ClickNodeAction = {
    nodeId: 'n1',
    type: ActionTypes.CLICK_NODE
  };

  const ClickNode2Action = {
    nodeId: 'n2',
    type: ActionTypes.CLICK_NODE
  };

  const ClickRelativeAction = {
    nodeId: 'rel1',
    type: ActionTypes.CLICK_RELATIVE
  };

  const ClickShowTopologyForNodeAction = {
    nodeId: 'rel1',
    topologyId: 'topo2',
    type: ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE
  };

  const ClickSubTopologyAction = {
    topologyId: 'topo1-grouped',
    type: ActionTypes.CLICK_TOPOLOGY
  };

  const ClickTopologyAction = {
    topologyId: 'topo1',
    type: ActionTypes.CLICK_TOPOLOGY
  };

  const ClickTopology2Action = {
    topologyId: 'topo2',
    type: ActionTypes.CLICK_TOPOLOGY
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
    delta: {
      add: [{
        adjacency: ['n1', 'n2'],
        id: 'n1'
      }, {
        id: 'n2'
      }]
    },
    type: ActionTypes.RECEIVE_NODES_DELTA
  };

  const ReceiveNodesDeltaUpdateAction = {
    delta: {
      remove: ['n2'],
      update: [{
        adjacency: ['n1'],
        id: 'n1'
      }]
    },
    type: ActionTypes.RECEIVE_NODES_DELTA
  };

  const ReceiveTopologiesAction = {
    topologies: [{
      name: 'Topo1',
      options: [{
        defaultValue: 'off',
        id: 'option1',
        options: [
          {value: 'on'},
          {value: 'off'}
        ]
      }],
      stats: {
        node_count: 1
      },
      sub_topologies: [{
        name: 'topo 1 grouped',
        url: '/topo1-grouped'
      }],
      url: '/topo1'
    }, {
      name: 'Topo2',
      stats: {
        node_count: 0
      },
      sub_topologies: [{
        name: 'topo 2 sub',
        url: '/topo2-sub'
      }],
      url: '/topo2'
    }],
    type: ActionTypes.RECEIVE_TOPOLOGIES
  };

  const ReceiveTopologiesHiddenAction = {
    topologies: [{
      name: 'Topo1',
      stats: {
        node_count: 1
      },
      url: '/topo1'
    }, {
      hide_if_empty: true,
      name: 'Topo2',
      stats: { filtered_nodes: 0, node_count: 0 },
      sub_topologies: [{
        hide_if_empty: true,
        name: 'topo 2 sub',
        stats: { filtered_nodes: 0, node_count: 0 },
        url: '/topo2-sub',
      }],
      url: '/topo2'
    }],
    type: ActionTypes.RECEIVE_TOPOLOGIES
  };

  const RouteAction = {
    state: {},
    type: ActionTypes.ROUTE_TOPOLOGY
  };

  const ChangeInstanceAction = {
    type: ActionTypes.CHANGE_INSTANCE
  };

  // Basic tests

  it('returns initial state', () => {
    const nextState = reducer(undefined, {});
    expect(is(nextState, initialState)).toBeTruthy();
  });

  // topology tests

  it('init with no topologies', () => {
    const nextState = reducer(undefined, {});
    expect(nextState.get('topologies').size).toBe(0);
    expect(nextState.get('currentTopology')).toBeFalsy();
  });

  it('get current topology', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopologyAction);

    expect(nextState.get('topologies').size).toBe(2);
    expect(nextState.get('currentTopology').get('name')).toBe('Topo1');
    expect(nextState.get('currentTopology').get('url')).toBe('/topo1');
    expect(nextState.get('currentTopology').get('options').first().get('id')).toEqual('option1');
    expect(nextState.getIn(['currentTopology', 'options']).toJS()).toEqual([{
      defaultValue: 'off',
      id: 'option1',
      options: [
        { value: 'on'},
        { value: 'off'}
      ],
      selectType: 'one'
    }]);
  });

  it('get sub-topology', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickSubTopologyAction);

    expect(nextState.get('topologies').size).toBe(2);
    expect(nextState.get('currentTopology').get('name')).toBe('topo 1 grouped');
    expect(nextState.get('currentTopology').get('url')).toBe('/topo1-grouped');
    expect(nextState.get('currentTopology').get('options')).toBeUndefined();
  });

  // topology options

  it('changes topology option', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopologyAction);

    // default options
    expect(activeTopologyOptionsSelector(nextState).has('option1')).toBeTruthy();
    expect(activeTopologyOptionsSelector(nextState).get('option1')).toBeInstanceOf(Array);
    expect(activeTopologyOptionsSelector(nextState).get('option1')).toEqual(['off']);
    expect(getUrlState(nextState).topologyOptions).toBeUndefined();

    // turn on
    nextState = reducer(nextState, ChangeTopologyOptionAction);
    expect(activeTopologyOptionsSelector(nextState).get('option1')).toEqual(['on']);
    expect(getUrlState(nextState).topologyOptions.topo1.option1).toEqual(['on']);

    // turn off
    nextState = reducer(nextState, ChangeTopologyOptionAction2);
    expect(activeTopologyOptionsSelector(nextState).get('option1')).toEqual(['off']);
    expect(getUrlState(nextState).topologyOptions).toBeUndefined();

    // sub-topology should retain main topo options
    nextState = reducer(nextState, ClickSubTopologyAction);
    expect(activeTopologyOptionsSelector(nextState).get('option1')).toEqual(['off']);
    expect(getUrlState(nextState).topologyOptions).toBeUndefined();

    // other topology w/o options dont return options, but keep in app state
    nextState = reducer(nextState, ClickTopology2Action);
    expect(activeTopologyOptionsSelector(nextState).size).toEqual(0);
    expect(getUrlState(nextState).topologyOptions).toBeUndefined();
  });

  it('adds/removes a topology option', () => {
    const addAction = {
      option: 'namespace',
      topologyId: 'services',
      type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
      value: ['default', 'scope'],
    };
    const removeAction = {
      option: 'namespace',
      topologyId: 'services',
      type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
      value: ['default']
    };
    let nextState = initialState;
    nextState = reducer(nextState, { topologies, type: ActionTypes.RECEIVE_TOPOLOGIES});
    nextState = reducer(nextState, { topologyId: 'services', type: ActionTypes.CLICK_TOPOLOGY });
    nextState = reducer(nextState, addAction);
    expect(activeTopologyOptionsSelector(nextState).toJS()).toEqual({
      namespace: ['default', 'scope'],
      pseudo: ['hide']
    });
    nextState = reducer(nextState, removeAction);
    expect(activeTopologyOptionsSelector(nextState).toJS()).toEqual({
      namespace: ['default'],
      pseudo: ['hide']
    });
  });

  it('sets topology options from route', () => {
    RouteAction.state = {
      selectedNodeId: null,
      topologyId: 'topo1',
      topologyOptions: {topo1: {option1: 'on'}}
    };

    let nextState = initialState;
    nextState = reducer(nextState, RouteAction);
    expect(activeTopologyOptionsSelector(nextState).get('option1')).toBe('on');
    expect(getUrlState(nextState).topologyOptions.topo1.option1).toBe('on');

    // stay same after topos have been received
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopologyAction);
    expect(activeTopologyOptionsSelector(nextState).get('option1')).toBe('on');
    expect(getUrlState(nextState).topologyOptions.topo1.option1).toBe('on');
  });

  it('uses default topology options from route', () => {
    RouteAction.state = {
      selectedNodeId: null,
      topologyId: 'topo1',
      topologyOptions: null
    };
    let nextState = initialState;
    nextState = reducer(nextState, RouteAction);
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopologyAction);
    expect(activeTopologyOptionsSelector(nextState).get('option1')).toEqual(['off']);
    expect(getUrlState(nextState).topologyOptions).toBeUndefined();
  });

  // nodes delta

  it('replaces adjacency on update', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    expect(nextState.get('nodes').toJS().n1.adjacency).toEqual(['n1', 'n2']);
    nextState = reducer(nextState, ReceiveNodesDeltaUpdateAction);
    expect(nextState.get('nodes').toJS().n1.adjacency).toEqual(['n1']);
  });

  // browsing

  it('shows nodes that were received', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    expect(nextState.get('nodes').toJS()).toEqual(NODE_SET);
  });

  it('knows a route was set', () => {
    let nextState = initialState;
    expect(nextState.get('routeSet')).toBeFalsy();
    nextState = reducer(nextState, RouteAction);
    expect(nextState.get('routeSet')).toBeTruthy();
  });

  it('gets selected node after click', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    nextState = reducer(nextState, ClickNodeAction);

    expect(nextState.get('selectedNodeId')).toBe('n1');
    expect(nextState.get('nodes').toJS()).toEqual(NODE_SET);

    nextState = reducer(nextState, deSelectNode);
    expect(nextState.get('selectedNodeId')).toBe(null);
    expect(nextState.get('nodes').toJS()).toEqual(NODE_SET);
  });

  it('keeps showing nodes on navigating back after node click', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopologyAction);
    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    expect(getUrlState(nextState).selectedNodeId).toBeUndefined();

    nextState = reducer(nextState, ClickNodeAction);
    expect(getUrlState(nextState).selectedNodeId).toEqual('n1');

    // go back in browsing
    RouteAction.state = {selectedNodeId: null, topologyId: 'topo1'};
    nextState = reducer(nextState, RouteAction);
    expect(nextState.get('selectedNodeId')).toBeNull();
    expect(nextState.get('nodes').toJS()).toEqual(NODE_SET);
  });

  it('closes details when changing topologies', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopologyAction);
    nextState = reducer(nextState, ReceiveNodesDeltaAction);

    expect(getUrlState(nextState).selectedNodeId).toBeUndefined();
    expect(getUrlState(nextState).topologyId).toEqual('topo1');

    nextState = reducer(nextState, ClickNodeAction);
    expect(getUrlState(nextState).selectedNodeId).toEqual('n1');
    expect(getUrlState(nextState).topologyId).toEqual('topo1');

    nextState = reducer(nextState, ClickSubTopologyAction);
    expect(getUrlState(nextState).selectedNodeId).toBeUndefined();
    expect(getUrlState(nextState).topologyId).toEqual('topo1-grouped');
  });

  // connection errors

  it('resets topology on websocket reconnect', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    expect(nextState.get('nodes').toJS()).toEqual(NODE_SET);

    nextState = reducer(nextState, CloseWebsocketAction);
    expect(nextState.get('websocketClosed')).toBeTruthy();
    // keep showing old nodes
    expect(nextState.get('nodes').toJS()).toEqual(NODE_SET);

    nextState = reducer(nextState, OpenWebsocketAction);
    expect(nextState.get('websocketClosed')).toBeFalsy();
  });

  // adjacency test

  it('returns the correct adjacency set for a node', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    expect(getAdjacentNodes(nextState).size).toEqual(0);

    nextState = reducer(nextState, ClickNodeAction);
    expect(getAdjacentNodes(nextState, 'n1').size).toEqual(2);
    expect(getAdjacentNodes(nextState, 'n1').has('n1')).toBeTruthy();
    expect(getAdjacentNodes(nextState, 'n1').has('n2')).toBeTruthy();

    nextState = reducer(nextState, deSelectNode);
    expect(getAdjacentNodes(nextState).size).toEqual(0);
  });

  // empty topology

  it('detects that the nodes display is empty', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopologyAction);
    expect(isNodesDisplayEmpty(nextState)).toBeTruthy();

    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    expect(isNodesDisplayEmpty(nextState)).toBeFalsy();

    nextState = reducer(nextState, ClickTopology2Action);
    expect(isNodesDisplayEmpty(nextState)).toBeTruthy();

    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    expect(isNodesDisplayEmpty(nextState)).toBeFalsy();
  });

  it('detects that the topo stats are empty', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopologyAction);
    expect(isTopologyNodeCountZero(nextState)).toBeFalsy();

    nextState = reducer(nextState, ReceiveNodesDeltaAction);
    expect(isTopologyNodeCountZero(nextState)).toBeFalsy();

    nextState = reducer(nextState, ClickTopology2Action);
    expect(isTopologyNodeCountZero(nextState)).toBeTruthy();

    nextState = reducer(nextState, ClickTopologyAction);
    expect(isTopologyNodeCountZero(nextState)).toBeFalsy();
  });

  it('keeps hidden topology visible if selected', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, ClickTopology2Action);
    nextState = reducer(nextState, ReceiveTopologiesHiddenAction);
    expect(nextState.get('currentTopologyId')).toEqual('topo2');
    expect(nextState.get('topologies').toJS().length).toEqual(2);
  });

  it('keeps hidden topology visible if sub_topology selected', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ReceiveTopologiesAction);
    nextState = reducer(nextState, { topologyId: 'topo2-sub', type: ActionTypes.CLICK_TOPOLOGY });
    nextState = reducer(nextState, ReceiveTopologiesHiddenAction);
    expect(nextState.get('currentTopologyId')).toEqual('topo2-sub');
    expect(nextState.get('topologies').toJS().length).toEqual(2);
  });

  it('hides hidden topology if not selected', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ClickTopologyAction);
    nextState = reducer(nextState, ReceiveTopologiesHiddenAction);
    expect(nextState.get('topologies').toJS().length).toEqual(1);
  });

  // selection of relatives

  it('keeps relatives as a stack', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ClickNodeAction);
    expect(nextState.get('selectedNodeId')).toBe('n1');
    expect(nextState.get('nodeDetails').size).toEqual(1);
    expect(nextState.get('nodeDetails').has('n1')).toBeTruthy();
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('n1');

    nextState = reducer(nextState, ClickRelativeAction);
    // stack relative, first node stays main node
    expect(nextState.get('selectedNodeId')).toBe('n1');
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('rel1');
    expect(nextState.get('nodeDetails').size).toEqual(2);
    expect(nextState.get('nodeDetails').has('rel1')).toBeTruthy();

    // click on first node should clear the stack
    nextState = reducer(nextState, ClickNodeAction);
    expect(nextState.get('selectedNodeId')).toBe('n1');
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('n1');
    expect(nextState.get('nodeDetails').size).toEqual(1);
    expect(nextState.get('nodeDetails').has('rel1')).toBeFalsy();
  });

  it('keeps clears stack when sibling is clicked', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ClickNodeAction);
    expect(nextState.get('selectedNodeId')).toBe('n1');
    expect(nextState.get('nodeDetails').size).toEqual(1);
    expect(nextState.get('nodeDetails').has('n1')).toBeTruthy();
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('n1');

    nextState = reducer(nextState, ClickRelativeAction);
    // stack relative, first node stays main node
    expect(nextState.get('selectedNodeId')).toBe('n1');
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('rel1');
    expect(nextState.get('nodeDetails').size).toEqual(2);
    expect(nextState.get('nodeDetails').has('rel1')).toBeTruthy();

    // click on sibling node should clear the stack
    nextState = reducer(nextState, ClickNode2Action);
    expect(nextState.get('selectedNodeId')).toBe('n2');
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('n2');
    expect(nextState.get('nodeDetails').size).toEqual(1);
    expect(nextState.get('nodeDetails').has('n1')).toBeFalsy();
    expect(nextState.get('nodeDetails').has('rel1')).toBeFalsy();
  });

  it('selectes relatives topology while keeping node selected', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ClickTopologyAction);
    nextState = reducer(nextState, ReceiveTopologiesAction);
    expect(nextState.get('currentTopology').get('name')).toBe('Topo1');

    nextState = reducer(nextState, ClickNodeAction);
    expect(nextState.get('selectedNodeId')).toBe('n1');
    expect(nextState.get('nodeDetails').size).toEqual(1);
    expect(nextState.get('nodeDetails').has('n1')).toBeTruthy();
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('n1');

    nextState = reducer(nextState, ClickRelativeAction);
    // stack relative, first node stays main node
    expect(nextState.get('selectedNodeId')).toBe('n1');
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('rel1');
    expect(nextState.get('nodeDetails').size).toEqual(2);
    expect(nextState.get('nodeDetails').has('rel1')).toBeTruthy();

    // click switches over to relative's topology and selectes relative
    nextState = reducer(nextState, ClickShowTopologyForNodeAction);
    expect(nextState.get('selectedNodeId')).toBe('rel1');
    expect(nextState.get('nodeDetails').keySeq().last()).toEqual('rel1');
    expect(nextState.get('nodeDetails').size).toEqual(1);
    expect(nextState.get('currentTopology').get('name')).toBe('Topo2');
  });
  it('closes the help dialog if the canvas is clicked', () => {
    let nextState = initialState.set('showingHelp', true);
    nextState = reducer(nextState, { type: ActionTypes.CLICK_BACKGROUND });
    expect(nextState.get('showingHelp')).toBe(false);
  });
  it('switches to table view when complexity is high', () => {
    let nextState = initialState.set('currentTopology', fromJS(topologies[0]));
    nextState = reducer(nextState, {type: ActionTypes.SET_RECEIVED_NODES_DELTA});
    expect(nextState.get('topologyViewMode')).toEqual(TABLE_VIEW_MODE);
    expect(nextState.get('initialNodesLoaded')).toBe(true);
  });
  it('cleans up old adjacencies', () => {
    // Add some nodes
    const action1 = {
      delta: { add: [{ id: 'n1' }, { id: 'n2' }] },
      type: ActionTypes.RECEIVE_NODES_DELTA
    };
    // Show nodes as connected
    const action2 = {
      delta: {
        update: [{ adjacency: ['n2'], id: 'n1' }]
      },
      type: ActionTypes.RECEIVE_NODES_DELTA
    };
    // Remove the connection
    const action3 = {
      delta: {
        update: [{ id: 'n1' }]
      },
      type: ActionTypes.RECEIVE_NODES_DELTA
    };
    let nextState = reducer(initialState, action1);
    nextState = reducer(nextState, action2);
    nextState = reducer(nextState, action3);
    expect(nextState.getIn(['nodes', 'n1', 'adjacency'])).toBeFalsy();
  });
  it('removes non-transferrable state values when changing instances', () => {
    let nextState = initialState;
    nextState = reducer(nextState, ClickNodeAction);
    expect(nextState.get('selectedNodeId')).toEqual('n1');
    expect(nextState.getIn(['nodeDetails', 'n1'])).toBeTruthy();
    nextState = reducer(nextState, ChangeInstanceAction);
    expect(nextState.get('selectedNodeId')).toBeFalsy();
    expect(nextState.getIn(['nodeDetails', 'n1'])).toBeFalsy();
  });
  it('highlights bidirectional edges', () => {
    const action = {
      edgeId: constructEdgeId('abc123', 'def456'),
      type: ActionTypes.ENTER_EDGE
    };
    const nextState = reducer(initialState, action);
    expect(highlightedEdgeIdsSelector(nextState).toJS()).toEqual([
      constructEdgeId('abc123', 'def456'),
      constructEdgeId('def456', 'abc123')
    ]);
  });
});
