import page from 'page';

import { route } from '../actions/app-actions';
import AppStore from '../stores/app-store';

export function updateRoute() {
  const state = AppStore.getAppState();
  const stateUrl = JSON.stringify(state);
  const dispatch = false;

  page.show('/state/' + stateUrl, state, dispatch);
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
