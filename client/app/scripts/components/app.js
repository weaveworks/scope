/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var Logo = require('./logo');
var SearchBar = require('./search-bar.js');
var AppStore = require('../stores/app-store');
var Topologies = require('./topologies.js');
var TopologyStore = require('../stores/topology-store');
var WebapiUtils = require('../utils/web-api-utils');
var AppActions = require('../actions/app-actions');
var Details = require('./details');
var Nodes = require('./nodes');
var ViewOptions = require('./view-options');
var RouterUtils = require('../utils/router-utils');


var ESC_KEY_CODE = 27;

function getStateFromStores() {
	return {
		selectedNodeId: AppStore.getSelectedNodeId(),
		nodeDetails: AppStore.getNodeDetails(),
		nodes: TopologyStore.getNodes(),
		topologies: AppStore.getTopologies(),
		activeTopology: AppStore.getCurrentTopology(),
		activeTopologyMode: AppStore.getCurrentTopologyMode()
	}
}


var App = React.createClass({

	getInitialState: function() {
		return getStateFromStores();
	},

	componentDidMount: function() {
		TopologyStore.on(TopologyStore.CHANGE_EVENT, this.onChange);
		AppStore.on(AppStore.CHANGE_EVENT, this.onChange);
		window.addEventListener('keyup', this.onKeyPress);

		RouterUtils.getRouter().start({hashbang: true});
		WebapiUtils.getTopologies();
	},

	onChange: function() {
		this.setState(getStateFromStores());
	},

	onKeyPress: function(ev) {
		if (ev.keyCode === ESC_KEY_CODE) {
			AppActions.hitEsc();
		}
	},

	render: function() {
		var showingDetails = this.state.selectedNodeId;

		return (
			<div>
				{showingDetails && <Details nodes={this.state.nodes}
					nodeId={this.state.selectedNodeId}
					details={this.state.nodeDetails}
					topology={this.state.activeTopology} /> }

				<div className="header">
					<div id="logo">
						<Logo />
					</div>
					<Topologies topologies={this.state.topologies} active={this.state.activeTopology} />
				</div>

				<Nodes nodes={this.state.nodes} />
			</div>
		);
	}

});

module.exports = App;
