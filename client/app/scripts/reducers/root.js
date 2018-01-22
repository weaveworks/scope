/* eslint-disable import/no-webpack-loader-syntax, import/no-unresolved */
import debug from 'debug';
import moment from 'moment';
import { size, each, includes, isEqual } from 'lodash';
import {
  fromJS,
  is as isDeepEqual,
  List as makeList,
  Map as makeMap,
  OrderedMap as makeOrderedMap,
} from 'immutable';

import ActionTypes from '../constants/action-types';
import {
  GRAPH_VIEW_MODE,
  TABLE_VIEW_MODE,
} from '../constants/naming';
import {
  graphExceedsComplexityThreshSelector,
  isResourceViewModeSelector,
} from '../selectors/topology';
import { isPausedSelector } from '../selectors/time-travel';
import { activeTopologyZoomCacheKeyPathSelector } from '../selectors/zooming';
import { applyPinnedSearches } from '../utils/search-utils';
import {
  findTopologyById,
  setTopologyUrlsById,
  updateTopologyIds,
  filterHiddenTopologies,
  addTopologyFullname,
  getDefaultTopology,
} from '../utils/topology-utils';

const log = debug('scope:app-store');
const error = debug('scope:error');

// Helpers

const topologySorter = topology => topology.get('rank');

// Initial values

export const initialState = makeMap({
  capabilities: makeMap(),
  contrastMode: false,
  controlPipes: makeOrderedMap(), // pipeId -> controlPipe
  controlStatus: makeMap(),
  currentTopology: null,
  currentTopologyId: null,
  errorUrl: null,
  exportingGraph: false,
  forceRelayout: false,
  gridSortedBy: null,
  gridSortedDesc: null,
  hostname: '...',
  hoveredMetricType: null,
  initialNodesLoaded: false,
  mouseOverEdgeId: null,
  mouseOverNodeId: null,
  nodeDetails: makeOrderedMap(), // nodeId -> details
  nodes: makeOrderedMap(), // nodeId -> node
  nodesLoaded: false,
  // nodes cache, infrequently updated, used for search & resource view
  nodesByTopology: makeMap(), // topologyId -> nodes
  // class of metric, e.g. 'cpu', rather than 'host_cpu' or 'process_cpu'.
  // allows us to keep the same metric "type" selected when the topology changes.
  pausedAt: null,
  pinnedMetricType: null,
  pinnedNetwork: null,
  plugins: makeList(),
  pinnedSearches: makeList(), // list of node filters
  routeSet: false,
  searchFocused: false,
  searchQuery: '',
  selectedNetwork: null,
  selectedNodeId: null,
  showingHelp: false,
  showingTimeTravel: false,
  showingTroubleshootingMenu: false,
  showingNetworks: false,
  timeTravelTransitioning: false,
  topologies: makeList(),
  topologiesLoaded: false,
  topologyOptions: makeOrderedMap(), // topologyId -> options
  topologyUrlsById: makeOrderedMap(), // topologyId -> topologyUrl
  topologyViewMode: GRAPH_VIEW_MODE,
  version: null,
  versionUpdate: null,
  // Set some initial numerical values to prevent NaN in case of edgy race conditions.
  viewport: makeMap({ width: 0, height: 0 }),
  websocketClosed: false,
  zoomCache: makeMap(),
  serviceImages: makeMap()
});

function calcSelectType(topology) {
  const result = {
    ...topology,
    options: topology.options && topology.options.map((option) => {
      // Server doesn't return the `selectType` key unless the option is something other than `one`.
      // Default to `one` if undefined, so the component doesn't have to handle this.
      option.selectType = option.selectType || 'one';
      return option;
    })
  };

  if (topology.sub_topologies) {
    result.sub_topologies = topology.sub_topologies.map(calcSelectType);
  }
  return result;
}

// adds ID field to topology (based on last part of URL path) and save urls in
// map for easy lookup
function processTopologies(state, nextTopologies) {
  // add IDs to topology objects in-place
  const topologiesWithId = updateTopologyIds(nextTopologies);
  // filter out hidden topos
  const visibleTopologies = filterHiddenTopologies(topologiesWithId, state.get('currentTopology'));
  // set `selectType` field for topology and sub_topologies options (recursive).
  const topologiesWithSelectType = visibleTopologies.map(calcSelectType);
  // cache URLs by ID
  state = state.set(
    'topologyUrlsById',
    setTopologyUrlsById(state.get('topologyUrlsById'), topologiesWithSelectType)
  );

  const topologiesWithFullnames = addTopologyFullname(topologiesWithSelectType);
  const immNextTopologies = fromJS(topologiesWithFullnames).sortBy(topologySorter);
  return state.set('topologies', immNextTopologies);
}

function setTopology(state, topologyId) {
  state = state.set('currentTopology', findTopologyById(state.get('topologies'), topologyId));
  return state.set('currentTopologyId', topologyId);
}

export function getDefaultTopologyOptions(state) {
  let topologyOptions = makeOrderedMap();
  state.get('topologies').forEach((topology) => {
    let defaultOptions = makeOrderedMap();
    if (topology.has('options') && topology.get('options')) {
      topology.get('options').forEach((option) => {
        const optionId = option.get('id');
        const defaultValue = option.get('defaultValue');
        defaultOptions = defaultOptions.set(optionId, [defaultValue]);
      });
    }
    if (defaultOptions.size) {
      topologyOptions = topologyOptions.set(topology.get('id'), defaultOptions);
    }
  });
  return topologyOptions;
}

function closeNodeDetails(state, nodeId) {
  const nodeDetails = state.get('nodeDetails');
  if (nodeDetails.size > 0) {
    const popNodeId = nodeId || nodeDetails.keySeq().last();
    // remove pipe if it belongs to the node being closed
    state = state.update(
      'controlPipes',
      controlPipes => controlPipes.filter(pipe => pipe.get('nodeId') !== popNodeId)
    );
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

function clearNodes(state) {
  return state
    .update('nodes', nodes => nodes.clear())
    .set('nodesLoaded', false);
}

// TODO: These state changes should probably be calculated from selectors.
function updateStateFromNodes(state) {
  // Apply pinned searches, filters nodes that dont match.
  state = applyPinnedSearches(state);

  // In case node or edge disappears before mouseleave event.
  const nodesIds = state.get('nodes').keySeq();
  if (!nodesIds.contains(state.get('mouseOverNodeId'))) {
    state = state.set('mouseOverNodeId', null);
  }
  if (!nodesIds.some(nodeId => includes(state.get('mouseOverEdgeId'), nodeId))) {
    state = state.set('mouseOverEdgeId', null);
  }

  // Update the nodes cache only if we're not in the resource view mode, as we
  // intentionally want to keep it static before we figure how to keep it up-to-date.
  if (!isResourceViewModeSelector(state)) {
    const nodesForCurrentTopologyKey = ['nodesByTopology', state.get('currentTopologyId')];
    state = state.setIn(nodesForCurrentTopologyKey, state.get('nodes'));
  }

  // Clear the error.
  state = state.set('errorUrl', null);

  return state;
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
      // set option on parent topology
      const topology = findTopologyById(state.get('topologies'), action.topologyId);
      if (topology) {
        const topologyId = topology.get('parentId') || topology.get('id');
        const optionKey = ['topologyOptions', topologyId, action.option];
        const currentOption = state.getIn(optionKey);

        if (!isEqual(currentOption, action.value)) {
          state = clearNodes(state);
        }

        state = state.setIn(optionKey, action.value);
      }
      return state;
    }

    case ActionTypes.SET_VIEWPORT_DIMENSIONS: {
      return state.mergeIn(['viewport'], {
        width: action.width,
        height: action.height,
      });
    }

    case ActionTypes.SET_EXPORTING_GRAPH: {
      return state.set('exportingGraph', action.exporting);
    }

    case ActionTypes.SORT_ORDER_CHANGED: {
      return state.merge({
        gridSortedBy: action.sortedBy,
        gridSortedDesc: action.sortedDesc,
      });
    }

    case ActionTypes.SET_VIEW_MODE: {
      return state.set('topologyViewMode', action.viewMode);
    }

    case ActionTypes.CACHE_ZOOM_STATE: {
      return state.setIn(activeTopologyZoomCacheKeyPathSelector(state), action.zoomState);
    }

    case ActionTypes.CLEAR_CONTROL_ERROR: {
      return state.removeIn(['controlStatus', action.nodeId, 'error']);
    }

    case ActionTypes.CLICK_BACKGROUND: {
      if (state.get('showingHelp')) {
        state = state.set('showingHelp', false);
      }

      if (state.get('showingTroubleshootingMenu')) {
        state = state.set('showingTroubleshootingMenu', false);
      }
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
        state = state.setIn(
          ['nodeDetails', action.nodeId],
          {
            id: action.nodeId,
            label: action.label,
            topologyId: action.topologyId || state.get('currentTopologyId'),
            origin,
          }
        );
        state = state.set('selectedNodeId', action.nodeId);
      }
      return state;
    }

    case ActionTypes.CLICK_RELATIVE: {
      if (state.hasIn(['nodeDetails', action.nodeId])) {
        // bring to front
        const details = state.getIn(['nodeDetails', action.nodeId]);
        state = state.deleteIn(['nodeDetails', action.nodeId]);
        state = state.setIn(['nodeDetails', action.nodeId], details);
      } else {
        state = state.setIn(
          ['nodeDetails', action.nodeId],
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

    case ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE: {
      state = state.update(
        'nodeDetails',
        nodeDetails => nodeDetails.filter((v, k) => k === action.nodeId)
      );
      state = state.update('controlPipes', controlPipes => controlPipes.clear());
      state = state.set('selectedNodeId', action.nodeId);

      if (action.topologyId !== state.get('currentTopologyId')) {
        state = setTopology(state, action.topologyId);
        state = clearNodes(state);
      }

      return state;
    }

    case ActionTypes.CLICK_TOPOLOGY: {
      state = closeAllNodeDetails(state);

      const currentTopologyId = state.get('currentTopologyId');
      if (action.topologyId !== currentTopologyId) {
        state = setTopology(state, action.topologyId);
        state = clearNodes(state);
      }

      return state;
    }

    //
    // time control
    //

    case ActionTypes.RESUME_TIME: {
      state = state.set('timeTravelTransitioning', true);
      state = state.set('showingTimeTravel', false);
      return state.set('pausedAt', null);
    }

    case ActionTypes.PAUSE_TIME_AT_NOW: {
      state = state.set('showingTimeTravel', false);
      state = state.set('timeTravelTransitioning', false);
      return state.set('pausedAt', moment().utc().format());
    }

    case ActionTypes.START_TIME_TRAVEL: {
      state = state.set('showingTimeTravel', true);
      state = state.set('timeTravelTransitioning', false);
      return state.set('pausedAt', action.timestamp || moment().utc().format());
    }

    case ActionTypes.JUMP_TO_TIME: {
      state = state.set('timeTravelTransitioning', true);
      return state.set('pausedAt', action.timestamp);
    }

    case ActionTypes.FINISH_TIME_TRAVEL_TRANSITION: {
      state = state.set('timeTravelTransitioning', false);
      return clearNodes(state);
    }

    //
    // websockets
    //

    case ActionTypes.OPEN_WEBSOCKET: {
      return state.set('websocketClosed', false);
    }

    case ActionTypes.CLOSE_WEBSOCKET: {
      return state.set('websocketClosed', true);
    }

    //
    // networks
    //

    case ActionTypes.SHOW_NETWORKS: {
      if (!action.visible) {
        state = state.set('selectedNetwork', null);
        state = state.set('pinnedNetwork', null);
      }
      return state.set('showingNetworks', action.visible);
    }

    case ActionTypes.SELECT_NETWORK: {
      return state.set('selectedNetwork', action.networkId);
    }

    case ActionTypes.PIN_NETWORK: {
      return state.merge({
        pinnedNetwork: action.networkId,
        selectedNetwork: action.networkId
      });
    }

    case ActionTypes.UNPIN_NETWORK: {
      return state.merge({
        pinnedNetwork: null,
      });
    }

    //
    // metrics
    //

    case ActionTypes.HOVER_METRIC: {
      return state.set('hoveredMetricType', action.metricType);
    }

    case ActionTypes.UNHOVER_METRIC: {
      return state.set('hoveredMetricType', null);
    }

    case ActionTypes.PIN_METRIC: {
      return state.set('pinnedMetricType', action.metricType);
    }

    case ActionTypes.UNPIN_METRIC: {
      return state.set('pinnedMetricType', null);
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
      return state.set('searchQuery', action.searchQuery);
    }

    case ActionTypes.ENTER_EDGE: {
      return state.set('mouseOverEdgeId', action.edgeId);
    }

    case ActionTypes.ENTER_NODE: {
      return state.set('mouseOverNodeId', action.nodeId);
    }

    case ActionTypes.LEAVE_EDGE: {
      return state.set('mouseOverEdgeId', null);
    }

    case ActionTypes.LEAVE_NODE: {
      return state.set('mouseOverNodeId', null);
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
      const pinnedSearches = state.get('pinnedSearches');
      state = state.setIn(['pinnedSearches', pinnedSearches.size], action.query);
      state = state.set('searchQuery', '');
      return applyPinnedSearches(state);
    }

    case ActionTypes.RECEIVE_CONTROL_NODE_REMOVED: {
      return closeNodeDetails(state, action.nodeId);
    }

    case ActionTypes.RECEIVE_CONTROL_PIPE: {
      return state.setIn(['controlPipes', action.pipeId], makeOrderedMap({
        id: action.pipeId,
        nodeId: action.nodeId,
        raw: action.rawTty,
        resizeTtyControl: action.resizeTtyControl,
        control: action.control
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
      // Ignore the update if paused and the timestamp didn't change.
      const setTimestamp = state.getIn(['nodeDetails', action.details.id, 'timestamp']);
      if (isPausedSelector(state) && action.requestTimestamp === setTimestamp) {
        return state;
      }

      state = state.set('errorUrl', null);

      // disregard if node is not selected anymore
      if (state.hasIn(['nodeDetails', action.details.id])) {
        state = state.updateIn(['nodeDetails', action.details.id], obj => ({
          ...obj,
          notFound: false,
          timestamp: action.requestTimestamp,
          details: action.details,
        }));
      }
      return state;
    }

    case ActionTypes.SET_RECEIVED_NODES_DELTA: {
      // Turn on the table view if the graph is too complex, but skip
      // this block if the user has already loaded topologies once.
      if (!state.get('initialNodesLoaded') && !state.get('nodesLoaded')) {
        if (state.get('topologyViewMode') === GRAPH_VIEW_MODE) {
          state = graphExceedsComplexityThreshSelector(state)
            ? state.set('topologyViewMode', TABLE_VIEW_MODE) : state;
        }
        state = state.set('initialNodesLoaded', true);
      }
      return state.set('nodesLoaded', true);
    }

    case ActionTypes.RECEIVE_NODES_DELTA: {
      // Ignore periodic nodes updates after the first load when paused.
      if (state.get('nodesLoaded') && state.get('pausedAt')) {
        return state;
      }

      log(
        'RECEIVE_NODES_DELTA',
        'remove', size(action.delta.remove),
        'update', size(action.delta.update),
        'add', size(action.delta.add),
        'reset', action.delta.reset
      );

      if (action.delta.reset) {
        state = state.set('nodes', makeMap());
      }

      // remove nodes that no longer exist
      each(action.delta.remove, (nodeId) => {
        state = state.deleteIn(['nodes', nodeId]);
      });

      // update existing nodes
      each(action.delta.update, (node) => {
        if (state.hasIn(['nodes', node.id])) {
          // TODO: Implement a manual deep update here, as it might bring a great benefit
          // to our nodes selectors (e.g. layout engine would be completely bypassed if the
          // adjacencies would stay the same but the metrics would get updated).
          state = state.setIn(['nodes', node.id], fromJS(node));
        }
      });

      // add new nodes
      each(action.delta.add, (node) => {
        state = state.setIn(['nodes', node.id], fromJS(node));
      });

      return updateStateFromNodes(state);
    }

    case ActionTypes.RECEIVE_NODES: {
      state = state.set('timeTravelTransitioning', false);
      state = state.set('nodes', fromJS(action.nodes));
      state = state.set('nodesLoaded', true);
      return updateStateFromNodes(state);
    }

    case ActionTypes.RECEIVE_NODES_FOR_TOPOLOGY: {
      return state.setIn(['nodesByTopology', action.topologyId], fromJS(action.nodes));
    }

    case ActionTypes.RECEIVE_NOT_FOUND: {
      if (state.hasIn(['nodeDetails', action.nodeId])) {
        state = state.updateIn(['nodeDetails', action.nodeId], obj => ({
          ...obj,
          timestamp: action.requestTimestamp,
          notFound: true,
        }));
      }
      return state;
    }

    case ActionTypes.RECEIVE_TOPOLOGIES: {
      state = state.set('errorUrl', null);
      state = state.update('topologyUrlsById', topologyUrlsById => topologyUrlsById.clear());
      state = processTopologies(state, action.topologies);
      const currentTopologyId = state.get('currentTopologyId');
      if (!currentTopologyId || !findTopologyById(state.get('topologies'), currentTopologyId)) {
        state = state.set('currentTopologyId', getDefaultTopology(state.get('topologies')));
        log(`Set currentTopologyId to ${state.get('currentTopologyId')}`);
      }
      state = setTopology(state, state.get('currentTopologyId'));
      // only set on first load, if options are not already set via route
      if (!state.get('topologiesLoaded') && state.get('topologyOptions').size === 0) {
        state = state.set('topologyOptions', getDefaultTopologyOptions(state));
      }
      state = state.set('topologiesLoaded', true);

      return state;
    }

    case ActionTypes.RECEIVE_API_DETAILS: {
      state = state.set('errorUrl', null);

      return state.merge({
        capabilities: action.capabilities,
        hostname: action.hostname,
        plugins: action.plugins,
        version: action.version,
        versionUpdate: action.newVersion,
      });
    }

    case ActionTypes.ROUTE_TOPOLOGY: {
      state = state.set('routeSet', true);
      state = state.set('pinnedSearches', makeList(action.state.pinnedSearches));
      state = state.set('searchQuery', action.state.searchQuery || '');
      if (state.get('currentTopologyId') !== action.state.topologyId) {
        state = clearNodes(state);
      }
      state = setTopology(state, action.state.topologyId);
      state = state.merge({
        selectedNodeId: action.state.selectedNodeId,
        pinnedMetricType: action.state.pinnedMetricType,
      });
      if (action.state.topologyViewMode) {
        state = state.set('topologyViewMode', action.state.topologyViewMode);
      }
      if (action.state.gridSortedBy) {
        state = state.set('gridSortedBy', action.state.gridSortedBy);
      }
      if (action.state.gridSortedDesc !== undefined) {
        state = state.set('gridSortedDesc', action.state.gridSortedDesc);
      }
      if (action.state.showingNetworks) {
        state = state.set('showingNetworks', action.state.showingNetworks);
      }
      if (action.state.pinnedNetwork) {
        state = state.set('pinnedNetwork', action.state.pinnedNetwork);
        state = state.set('selectedNetwork', action.state.pinnedNetwork);
      }
      if (action.state.controlPipe) {
        state = state.set('controlPipes', makeOrderedMap({
          [action.state.controlPipe.id]:
            makeOrderedMap(action.state.controlPipe)
        }));
      } else {
        state = state.update('controlPipes', controlPipes => controlPipes.clear());
      }
      if (action.state.nodeDetails) {
        const actionNodeDetails = makeOrderedMap(action.state.nodeDetails.map(h => [h.id, h]));
        // check if detail IDs have changed
        if (!isDeepEqual(state.get('nodeDetails').keySeq(), actionNodeDetails.keySeq())) {
          state = state.set('nodeDetails', actionNodeDetails);
        }
      } else {
        state = state.update('nodeDetails', nodeDetails => nodeDetails.clear());
      }
      // Use the default topology options for all the fields not
      // explicitly listed in the Scope state (URL or local storage).
      state = state.set(
        'topologyOptions',
        getDefaultTopologyOptions(state).mergeDeep(action.state.topologyOptions),
      );
      return state;
    }

    case ActionTypes.UNPIN_SEARCH: {
      const pinnedSearches = state.get('pinnedSearches').filter(query => query !== action.query);
      state = state.set('pinnedSearches', pinnedSearches);
      return applyPinnedSearches(state);
    }

    case ActionTypes.DEBUG_TOOLBAR_INTERFERING: {
      return action.fn(state);
    }

    case ActionTypes.TOGGLE_TROUBLESHOOTING_MENU: {
      return state.set('showingTroubleshootingMenu', !state.get('showingTroubleshootingMenu'));
    }

    case ActionTypes.CHANGE_INSTANCE: {
      state = closeAllNodeDetails(state);
      return state;
    }

    case ActionTypes.TOGGLE_CONTRAST_MODE: {
      return state.set('contrastMode', action.enabled);
    }

    case ActionTypes.SHUTDOWN: {
      return clearNodes(state);
    }

    case ActionTypes.REQUEST_SERVICE_IMAGES: {
      return state.setIn(['serviceImages', action.serviceId], {
        isFetching: true
      });
    }

    case ActionTypes.RECEIVE_SERVICE_IMAGES: {
      const { service, errors, serviceId } = action;

      return state.setIn(['serviceImages', serviceId], {
        isFetching: false,
        containers: service ? service.Containers : null,
        errors
      });
    }

    default: {
      return state;
    }
  }
}

export default rootReducer;
