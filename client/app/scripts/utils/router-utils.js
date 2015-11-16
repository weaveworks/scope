import page from 'page';

import { route } from '../actions/app-actions';
import AppStore from '../stores/app-store';

//
// TODO: move this logic somewhere else.
//
function shouldReplaceState(prevState, nextState) {
  return (
    // Opening a new terminal while an existing one is open.
    (prevState.controlPipe && nextState.controlPipe) ||
    // Closing a terminal.
    (prevState.controlPipe && !nextState.controlPipe)
  );
}

export function updateRoute() {
  const state = AppStore.getAppState();
  const stateUrl = JSON.stringify(state);
  const dispatch = false;
  const urlStateString = window.location.hash.replace('#!/state/', '') || '{}';
  const prevState = JSON.parse(urlStateString);

  if (shouldReplaceState(prevState, state)) {
    // Replace the top of the history rather than pushing on a new item.
    page.replace('/state/' + stateUrl, state, dispatch);
  } else {
    page.show('/state/' + stateUrl, state, dispatch);
  }
}

page('/', function() {
  updateRoute();
});

page('/state/:state', function(ctx) {
  const state = JSON.parse(ctx.params.state);
  route(state);
});

export function getRouter() {
  return page;
}
