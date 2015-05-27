const page = require('page');

const AppActions = require('../actions/app-actions');
const AppStore = require('../stores/app-store');

function updateRoute() {
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
  AppActions.route(state);
});

module.exports = {
  getRouter: function() {
    return page;
  },

  updateRoute: updateRoute
};
