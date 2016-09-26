import page from 'page';

import { route } from '../actions/app-actions';
import { storageGet, storageSet } from './storage-utils';

//
// page.js won't match the routes below if ":state" has a slash in it, so replace those before we
// load the state into the URL.
//
const SLASH = '/';
const SLASH_REPLACEMENT = '<SLASH>';
const PERCENT = '%';
const PERCENT_REPLACEMENT = '<PERCENT>';
const STORAGE_STATE_KEY = 'scopeViewState';

function encodeURL(url) {
  return url
    .replace(new RegExp(PERCENT, 'g'), PERCENT_REPLACEMENT)
    .replace(new RegExp(SLASH, 'g'), SLASH_REPLACEMENT);
}

function decodeURL(url) {
  return decodeURIComponent(url.replace(new RegExp(SLASH_REPLACEMENT, 'g'), SLASH))
    .replace(new RegExp(PERCENT_REPLACEMENT, 'g'), PERCENT);
}

function shouldReplaceState(prevState, nextState) {
  // Opening a new terminal while an existing one is open.
  const terminalToTerminal = (prevState.controlPipe && nextState.controlPipe);
  // Closing a terminal.
  const closingTheTerminal = (prevState.controlPipe && !nextState.controlPipe);

  return terminalToTerminal || closingTheTerminal;
}

export function getUrlState(state) {
  const cp = state.get('controlPipes').last();
  const nodeDetails = state.get('nodeDetails').toIndexedSeq().map(details => ({
    id: details.id, label: details.label, topologyId: details.topologyId
  }));

  const urlState = {
    controlPipe: cp ? cp.toJS() : null,
    topologyViewMode: state.get('gridMode') ? 'grid' : 'topo',
    nodeDetails: nodeDetails.toJS(),
    pinnedMetricType: state.get('pinnedMetricType'),
    pinnedSearches: state.get('pinnedSearches').toJS(),
    searchQuery: state.get('searchQuery'),
    selectedNodeId: state.get('selectedNodeId'),
    gridSortBy: state.get('gridSortBy'),
    gridSortedDesc: state.get('gridSortedDesc'),
    topologyId: state.get('currentTopologyId'),
    topologyOptions: state.get('topologyOptions').toJS() // all options
  };

  if (state.get('showingNetworks')) {
    urlState.showingNetworks = true;
    if (state.get('pinnedNetwork')) {
      urlState.pinnedNetwork = state.get('pinnedNetwork');
    }
  }

  return urlState;
}

export function updateRoute(getState) {
  const state = getUrlState(getState());
  const stateUrl = encodeURL(JSON.stringify(state));
  const dispatch = false;
  const urlStateString = window.location.hash
    .replace('#!/state/', '')
    .replace('#!/', '') || '{}';
  const prevState = JSON.parse(decodeURL(urlStateString));

  // back up state in storage as well
  storageSet(STORAGE_STATE_KEY, stateUrl);

  if (shouldReplaceState(prevState, state)) {
    // Replace the top of the history rather than pushing on a new item.
    page.replace(`/state/${stateUrl}`, state, dispatch);
  } else {
    page.show(`/state/${stateUrl}`, state, dispatch);
  }
}


export function getRouter(dispatch, initialState) {
  // strip any trailing '/'s.
  page.base(window.location.pathname.replace(/\/$/, ''));

  page('/', () => {
    // recover from storage state on empty URL
    const storageState = storageGet(STORAGE_STATE_KEY);
    if (storageState) {
      // push storage state to URL
      window.location.hash = `!/state/${storageState}`;
      const parsedState = JSON.parse(decodeURL(storageState));
      const mergedState = Object.assign(initialState, parsedState);
      dispatch(route(mergedState));
    } else {
      dispatch(route(initialState));
    }
  });

  page('/state/:state', (ctx) => {
    const state = JSON.parse(decodeURL(ctx.params.state));
    dispatch(route(state));
  });

  return page;
}
