import _ from 'lodash';
import debug from 'debug';
import { fromJS, is as isDeepEqual, List as makeList, Map as makeMap,
  OrderedMap as makeOrderedMap, Set as makeSet } from 'immutable';

import ActionTypes from '../constants/action-types';
import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { applyPinnedSearches, updateNodeMatches } from '../utils/search-utils';
import { findTopologyById, getAdjacentNodes, setTopologyUrlsById,
  updateTopologyIds, filterHiddenTopologies } from '../utils/topology-utils';

const log = debug('scope:app-store');
const error = debug('scope:error');

// Helpers

const topologySorter = topology => topology.get('rank');

// Initial values

export const initialState = makeMap({
  availableCanvasMetrics: makeList(),
  controlPipes: makeOrderedMap(), // pipeId -> controlPipe
  controlStatus: makeMap(),
  currentTopology: null,
  currentTopologyId: 'containers',
  errorUrl: null,
  forceRelayout: false,
  highlightedEdgeIds: makeSet(),
  highlightedNodeIds: makeSet(),
  hostname: '...',
  mouseOverEdgeId: null,
  mouseOverNodeId: null,
  nodeDetails: makeOrderedMap(), // nodeId -> details
  nodes: makeOrderedMap(), // nodeId -> node
  // nodes cache, infrequently updated, used for search
  nodesByTopology: makeMap(), // topologyId -> nodes
  pinnedMetric: null,
  // class of metric, e.g. 'cpu', rather than 'host_cpu' or 'process_cpu'.
  // allows us to keep the same metric "type" selected when the topology changes.
  pinnedMetricType: null,
  plugins: makeList(),
  pinnedSearches: makeList(), // list of node filters
  routeSet: false,
  searchFocused: false,
  searchNodeMatches: makeMap(),
  searchQuery: null,
  selectedMetric: null,
  selectedNodeId: null,
  showingHelp: false,
  topologies: makeList(),
  topologiesLoaded: false,
  topologyOptions: makeOrderedMap(), // topologyId -> options
  topologyUrlsById: makeOrderedMap(), // topologyId -> topologyUrl
  updatePausedAt: null, // Date
  version: '...',
  versionUpdate: null,
  websocketClosed: true,
  exportingGraph: false
});

// adds ID field to topology (based on last part of URL path) and save urls in
// map for easy lookup
function processTopologies(state, nextTopologies) {
  // filter out hidden topos
  const visibleTopologies = filterHiddenTopologies(nextTopologies);

  // add IDs to topology objects in-place
  const topologiesWithId = updateTopologyIds(visibleTopologies);

  // cache URLs by ID
  state = state.set('topologyUrlsById',
    setTopologyUrlsById(state.get('topologyUrlsById'), topologiesWithId));

  const immNextTopologies = fromJS(topologiesWithId).sortBy(topologySorter);
  return state.mergeDeepIn(['topologies'], immNextTopologies);
}

function setTopology(state, topologyId) {
  state = state.set('currentTopology', findTopologyById(
    state.get('topologies'), topologyId));
  return state.set('currentTopologyId', topologyId);
}

function setDefaultTopologyOptions(state, topologyList) {
  topologyList.forEach(topology => {
    let defaultOptions = makeOrderedMap();
    if (topology.has('options') && topology.get('options')) {
      topology.get('options').forEach((option) => {
        const optionId = option.get('id');
        const defaultValue = option.get('defaultValue');
        defaultOptions = defaultOptions.set(optionId, defaultValue);
      });
    }

    if (defaultOptions.size) {
      state = state.setIn(['topologyOptions', topology.get('id')],
        defaultOptions
      );
    }
  });
  return state;
}

function closeNodeDetails(state, nodeId) {
  const nodeDetails = state.get('nodeDetails');
  if (nodeDetails.size > 0) {
    const popNodeId = nodeId || nodeDetails.keySeq().last();
    // remove pipe if it belongs to the node being closed
    state = state.update('controlPipes',
      controlPipes => controlPipes.filter(pipe => pipe.get('nodeId') !== popNodeId));
    state = state.deleteIn(['nodeDetails', popNodeId]);
  }
  if (state.get('nodeDetails').size === 0 || state.get('selectedNodeId') === nodeId) {
    state = state.set('selectedNodeId', null);
  }
  return state;
}

function closeAllNodeDetails(state) {
  while (state.get('nodeDetails').size) {
    state = closeNodeDetails(state);
  }
  return state;
}

function resumeUpdate(state) {
  return state.set('updatePausedAt', null);
}

export function rootReducer(state = initialState, action) {
  if (!action.type) {
    error('Payload missing a type!', action);
  }

  switch (action.type) {
    case ActionTypes.BLUR_SEARCH: {
      return state.set('searchFocused', false);
    }

    case ActionTypes.CHANGE_TOPOLOGY_OPTION: {
      state = resumeUpdate(state);
      // set option on parent topology
      const topology = findTopologyById(state.get('topologies'), action.topologyId);
      if (topology) {
        const topologyId = topology.get('parentId') || topology.get('id');
        if (state.getIn(['topologyOptions', topologyId, action.option]) !== action.value) {
          state = state.update('nodes', nodes => nodes.clear());
        }
        state = state.setIn(
          ['topologyOptions', topologyId, action.option],
          action.value
        );
      }
      return state;
    }

    case ActionTypes.SET_EXPORTING_GRAPH: {
      return state.set('exportingGraph', action.exporting);
    }

    case ActionTypes.CLEAR_CONTROL_ERROR: {
      return state.removeIn(['controlStatus', action.nodeId, 'error']);
    }

    case ActionTypes.CLICK_BACKGROUND: {
      return closeAllNodeDetails(state);
    }

    case ActionTypes.CLICK_CLOSE_DETAILS: {
      return closeNodeDetails(state, action.nodeId);
    }

    case ActionTypes.CLICK_CLOSE_TERMINAL: {
      return state.update('controlPipes', controlPipes => controlPipes.clear());
    }

    case ActionTypes.CLICK_FORCE_RELAYOUT: {
      return state.set('forceRelayout', action.forceRelayout);
    }

    case ActionTypes.CLICK_NODE: {
      const prevSelectedNodeId = state.get('selectedNodeId');
      const prevDetailsStackSize = state.get('nodeDetails').size;

      // click on sibling closes all
      state = closeAllNodeDetails(state);

      // select new node if it's not the same (in that case just delesect)
      if (prevDetailsStackSize > 1 || prevSelectedNodeId !== action.nodeId) {
        // dont set origin if a node was already selected, suppresses animation
        const origin = prevSelectedNodeId === null ? action.origin : null;
        state = state.setIn(['nodeDetails', action.nodeId],
          {
            id: action.nodeId,
            label: action.label,
            origin,
            topologyId: state.get('currentTopologyId')
          }
        );
        state = state.set('selectedNodeId', action.nodeId);
      }
      return state;
    }

    case ActionTypes.CLICK_PAUSE_UPDATE: {
      return state.set('updatePausedAt', new Date);
    }

    case ActionTypes.CLICK_RELATIVE: {
      if (state.hasIn(['nodeDetails', action.nodeId])) {
        // bring to front
        const details = state.getIn(['nodeDetails', action.nodeId]);
        state = state.deleteIn(['nodeDetails', action.nodeId]);
        state = state.setIn(['nodeDetails', action.nodeId], details);
      } else {
        state = state.setIn(['nodeDetails', action.nodeId],
          {
            id: action.nodeId,
            label: action.label,
            origin: action.origin,
            topologyId: action.topologyId
          }
        );
      }
      return state;
    }

    case ActionTypes.CLICK_RESUME_UPDATE: {
      return resumeUpdate(state);
    }

    case ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE: {
      state = resumeUpdate(state);

      state = state.update('nodeDetails',
        nodeDetails => nodeDetails.filter((v, k) => k === action.nodeId));
      state = state.update('controlPipes', controlPipes => controlPipes.clear());
      state = state.set('selectedNodeId', action.nodeId);

      if (action.topologyId !== state.get('currentTopologyId')) {
        state = setTopology(state, action.topologyId);
        state = state.update('nodes', nodes => nodes.clear());
      }
      state = state.set('availableCanvasMetrics', makeList());
      return state;
    }

    case ActionTypes.CLICK_TOPOLOGY: {
      state = resumeUpdate(state);
      state = closeAllNodeDetails(state);

      if (action.topologyId !== state.get('currentTopologyId')) {
        state = setTopology(state, action.topologyId);
        state = state.update('nodes', nodes => nodes.clear());
      }
      state = state.set('availableCanvasMetrics', makeList());
      return state;
    }

    case ActionTypes.CLOSE_WEBSOCKET: {
      if (!state.get('websocketClosed')) {
        state = state.set('websocketClosed', true);
      }
      return state;
    }

    case ActionTypes.SELECT_METRIC: {
      return state.set('selectedMetric', action.metricId);
    }

    case ActionTypes.PIN_METRIC: {
      const metricTypes = makeMap(
        state.get('availableCanvasMetrics').map(m => [m.get('id'), m.get('label')]));
      return state.merge({
        pinnedMetric: action.metricId,
        pinnedMetricType: metricTypes.get(action.metricId),
        selectedMetric: action.metricId
      });
    }

    case ActionTypes.UNPIN_METRIC: {
      return state.merge({
        pinnedMetric: null,
        pinnedMetricType: null
      });
    }

    case ActionTypes.SHOW_HELP: {
      return state.set('showingHelp', true);
    }

    case ActionTypes.HIDE_HELP: {
      return state.set('showingHelp', false);
    }

    case ActionTypes.DESELECT_NODE: {
      return closeNodeDetails(state);
    }

    case ActionTypes.DO_CONTROL: {
      return state.setIn(['controlStatus', action.nodeId], makeMap({
        pending: true,
        error: null
      }));
    }

    case ActionTypes.DO_SEARCH: {
      state = state.set('searchQuery', action.searchQuery);
      return updateNodeMatches(state);
    }

    case ActionTypes.ENTER_EDGE: {
      // highlight adjacent nodes
      state = state.update('highlightedNodeIds', highlightedNodeIds => {
        highlightedNodeIds = highlightedNodeIds.clear();
        return highlightedNodeIds.union(action.edgeId.split(EDGE_ID_SEPARATOR));
      });

      // highlight edge
      state = state.update('highlightedEdgeIds', highlightedEdgeIds => {
        highlightedEdgeIds = highlightedEdgeIds.clear();
        return highlightedEdgeIds.add(action.edgeId);
      });

      return state;
    }

    case ActionTypes.ENTER_NODE: {
      const nodeId = action.nodeId;
      const adjacentNodes = getAdjacentNodes(state, nodeId);

      state = state.set('mouseOverNodeId', nodeId);

      // highlight adjacent nodes
      state = state.update('highlightedNodeIds', highlightedNodeIds => {
        highlightedNodeIds = highlightedNodeIds.clear();
        highlightedNodeIds = highlightedNodeIds.add(nodeId);
        return highlightedNodeIds.union(adjacentNodes);
      });

      // highlight edge
      state = state.update('highlightedEdgeIds', highlightedEdgeIds => {
        highlightedEdgeIds = highlightedEdgeIds.clear();
        if (adjacentNodes.size > 0) {
          // all neighbour combinations because we dont know which direction exists
          highlightedEdgeIds = highlightedEdgeIds.union(adjacentNodes.flatMap((adjacentId) => [
            [adjacentId, nodeId].join(EDGE_ID_SEPARATOR),
            [nodeId, adjacentId].join(EDGE_ID_SEPARATOR)
          ]));
        }
        return highlightedEdgeIds;
      });

      return state;
    }

    case ActionTypes.LEAVE_EDGE: {
      state = state.update('highlightedEdgeIds', highlightedEdgeIds => highlightedEdgeIds.clear());
      state = state.update('highlightedNodeIds', highlightedNodeIds => highlightedNodeIds.clear());
      return state;
    }

    case ActionTypes.LEAVE_NODE: {
      state = state.set('mouseOverNodeId', null);
      state = state.update('highlightedEdgeIds', highlightedEdgeIds => highlightedEdgeIds.clear());
      state = state.update('highlightedNodeIds', highlightedNodeIds => highlightedNodeIds.clear());
      return state;
    }

    case ActionTypes.OPEN_WEBSOCKET: {
      // flush nodes cache after re-connect
      state = state.update('nodes', nodes => nodes.clear());
      state = state.set('websocketClosed', false);
      return state;
    }

    case ActionTypes.DO_CONTROL_ERROR: {
      return state.setIn(['controlStatus', action.nodeId], makeMap({
        pending: false,
        error: action.error
      }));
    }
    case ActionTypes.DO_CONTROL_SUCCESS: {
      return state.setIn(['controlStatus', action.nodeId], makeMap({
        pending: false,
        error: null
      }));
    }

    case ActionTypes.FOCUS_SEARCH: {
      return state.set('searchFocused', true);
    }

    case ActionTypes.PIN_SEARCH: {
      state = state.set('searchQuery', '');
      state = updateNodeMatches(state);
      const pinnedSearches = state.get('pinnedSearches');
      state = state.setIn(['pinnedSearches', pinnedSearches.size], action.query);
      return applyPinnedSearches(state);
    }

    case ActionTypes.RECEIVE_CONTROL_NODE_REMOVED: {
      return closeNodeDetails(state, action.nodeId);
    }

    case ActionTypes.RECEIVE_CONTROL_PIPE: {
      return state.setIn(['controlPipes', action.pipeId], makeOrderedMap({
        id: action.pipeId,
        nodeId: action.nodeId,
        raw: action.rawTty
      }));
    }

    case ActionTypes.RECEIVE_CONTROL_PIPE_STATUS: {
      if (state.hasIn(['controlPipes', action.pipeId])) {
        state = state.setIn(['controlPipes', action.pipeId, 'status'], action.status);
      }
      return state;
    }

    case ActionTypes.RECEIVE_ERROR: {
      if (state.get('errorUrl') !== null) {
        state = state.set('errorUrl', action.errorUrl);
      }
      return state;
    }

    case ActionTypes.RECEIVE_NODE_DETAILS: {
      state = state.set('errorUrl', null);

      // disregard if node is not selected anymore
      if (state.hasIn(['nodeDetails', action.details.id])) {
        state = state.updateIn(['nodeDetails', action.details.id], obj => {
          const result = Object.assign({}, obj);
          result.notFound = false;
          result.details = action.details;
          return result;
        });
      }
      return state;
    }

    case ActionTypes.RECEIVE_NODES_DELTA: {
      const emptyMessage = !action.delta.add && !action.delta.remove
        && !action.delta.update;

      if (!emptyMessage) {
        log('RECEIVE_NODES_DELTA',
          'remove', _.size(action.delta.remove),
          'update', _.size(action.delta.update),
          'add', _.size(action.delta.add));
      }

      state = state.set('errorUrl', null);

      // nodes that no longer exist
      _.each(action.delta.remove, (nodeId) => {
        // in case node disappears before mouseleave event
        if (state.get('mouseOverNodeId') === nodeId) {
          state = state.set('mouseOverNodeId', null);
        }
        if (state.hasIn(['nodes', nodeId]) && _.includes(state.get('mouseOverEdgeId'), nodeId)) {
          state = state.set('mouseOverEdgeId', null);
        }
        state = state.deleteIn(['nodes', nodeId]);
      });

      // update existing nodes
      _.each(action.delta.update, (node) => {
        if (state.hasIn(['nodes', node.id])) {
          state = state.updateIn(['nodes', node.id], n => n.merge(fromJS(node)));
        }
      });

      // add new nodes
      _.each(action.delta.add, (node) => {
        state = state.setIn(['nodes', node.id], fromJS(node));
      });

      // apply pinned searches, filters nodes that dont match
      state = applyPinnedSearches(state);

      state = state.set('availableCanvasMetrics', state.get('nodes')
        .valueSeq()
        .flatMap(n => (n.get('metrics') || makeList()).map(m => (
          makeMap({id: m.get('id'), label: m.get('label')})
        )))
        .toSet()
        .toList()
        .sortBy(m => m.get('label')));

      const similarTypeMetric = state.get('availableCanvasMetrics')
        .find(m => m.get('label') === state.get('pinnedMetricType'));
      state = state.set('pinnedMetric', similarTypeMetric && similarTypeMetric.get('id'));
      // if something in the current topo is not already selected, select it.
      if (!state.get('availableCanvasMetrics')
        .map(m => m.get('id'))
        .toSet()
        .has(state.get('selectedMetric'))) {
        state = state.set('selectedMetric', state.get('pinnedMetric'));
      }

      // update nodes cache and search results
      state = state.setIn(['nodesByTopology', state.get('currentTopologyId')], state.get('nodes'));
      state = updateNodeMatches(state);

      return state;
    }

    case ActionTypes.RECEIVE_NODES_FOR_TOPOLOGY: {
      // not sure if mergeDeep() brings any benefit here
      state = state.setIn(['nodesByTopology', action.topologyId], fromJS(action.nodes));
      state = updateNodeMatches(state);
      return state;
    }

    case ActionTypes.RECEIVE_NOT_FOUND: {
      if (state.hasIn(['nodeDetails', action.nodeId])) {
        state = state.updateIn(['nodeDetails', action.nodeId], obj => {
          const result = Object.assign({}, obj);
          result.notFound = true;
          return result;
        });
      }
      return state;
    }

    case ActionTypes.RECEIVE_TOPOLOGIES: {
      state = state.set('errorUrl', null);
      state = state.update('topologyUrlsById', topologyUrlsById => topologyUrlsById.clear());
      state = processTopologies(state, action.topologies);
      state = setTopology(state, state.get('currentTopologyId'));
      // only set on first load, if options are not already set via route
      if (!state.get('topologiesLoaded') && state.get('topologyOptions').size === 0) {
        state = setDefaultTopologyOptions(state, state.get('topologies'));
      }
      state = state.set('topologiesLoaded', true);

      return state;
    }

    case ActionTypes.RECEIVE_API_DETAILS: {
      state = state.set('errorUrl', null);

      return state.merge({
        hostname: action.hostname,
        version: action.version,
        plugins: action.plugins,
        versionUpdate: action.newVersion
      });
    }

    case ActionTypes.ROUTE_TOPOLOGY: {
      state = state.set('routeSet', true);
      state = state.set('pinnedSearches', makeList(action.state.pinnedSearches));
      state = state.set('searchQuery', action.state.searchQuery || '');
      if (state.get('currentTopologyId') !== action.state.topologyId) {
        state = state.update('nodes', nodes => nodes.clear());
      }
      state = setTopology(state, action.state.topologyId);
      state = setDefaultTopologyOptions(state, state.get('topologies'));
      state = state.merge({
        selectedNodeId: action.state.selectedNodeId,
        pinnedMetricType: action.state.pinnedMetricType
      });
      if (action.state.controlPipe) {
        state = state.set('controlPipes', makeOrderedMap({
          [action.state.controlPipe.id]:
            makeOrderedMap(action.state.controlPipe)
        }));
      } else {
        state = state.update('controlPipes', controlPipes => controlPipes.clear());
      }
      if (action.state.nodeDetails) {
        const actionNodeDetails = makeOrderedMap(
          action.state.nodeDetails.map(obj => [obj.id, obj]));
        // check if detail IDs have changed
        if (!isDeepEqual(state.get('nodeDetails').keySeq(), actionNodeDetails.keySeq())) {
          state = state.set('nodeDetails', actionNodeDetails);
        }
      } else {
        state = state.update('nodeDetails', nodeDetails => nodeDetails.clear());
      }
      state = state.set('topologyOptions',
        fromJS(action.state.topologyOptions) || state.get('topologyOptions'));
      return state;
    }

    case ActionTypes.UNPIN_SEARCH: {
      const pinnedSearches = state.get('pinnedSearches').filter(query => query !== action.query);
      state = state.set('pinnedSearches', pinnedSearches);
      return applyPinnedSearches(state);
    }

    default: {
      return state;
    }
  }
}

export default rootReducer;
