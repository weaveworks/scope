
var page = require('page');

var AppActions = require('../actions/app-actions');
var AppStore = require('../stores/app-store');

page('/', function(ctx) {
	updateRoute();
});

page('/state/:state', function(ctx) {
	var state = JSON.parse(ctx.params.state);
	AppActions.route(state);
});

function updateRoute() {
	var state = AppStore.getAppState();
	var stateUrl = JSON.stringify(state);
	var dispatch = false;

	page.show('/state/' + stateUrl, state, dispatch);
}


module.exports = {
	getRouter: function() {
		return page;
	},

	updateRoute: updateRoute
};