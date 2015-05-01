/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var Logo = require('./logo');
var Adjacency = require('./adjacency.js');
var SearchBar = require('./search-bar.js');
var AppStore = require('../stores/app-store');
var Topologies = require('./topologies.js');
var Views = require('./views.js');
var TopologyStore = require('../stores/topology-store');
var WebapiUtils = require('../utils/web-api-utils');
var AppActions = require('../actions/app-actions');
var Details = require('./details');
var Chord = require('./chord');
var Matrix = require('./matrix');
var Nodes = require('./nodes');
var ViewOptions = require('./view-options');
var RouterUtils = require('../utils/router-utils');


var ESC_KEY_CODE = 27;

function getStateFromStores() {
	return {
		detailsView: AppStore.getDetailsView(),
		explorerExpandedNodes: AppStore.getExplorerExpandedNodes(),
		highlightedAdjacents: TopologyStore.getHighlightedAdjacents(),
		nodeDetails: AppStore.getNodeDetails(),
		nodes: TopologyStore.getNodes(),
		topologies: AppStore.getTopologies(),
		filterAdjacent: TopologyStore.getFilterAdjacent(),
		filterText: TopologyStore.getFilterText(),
		activeTopology: AppStore.getCurrentTopology(),
		activeTopologyMode: AppStore.getCurrentTopologyMode(),
		currentView: AppStore.getCurrentView()
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
		var showingAdjacency = this.state.currentView === 'adjacency';
		var showingMatrix = this.state.currentView === 'matrix';
		var showingChord = this.state.currentView === 'chord';
		var showingNodes = this.state.currentView === 'nodes';
		var showingDetails = this.state.detailsView;

		return (
			<div>
				{showingDetails && <Details nodes={this.state.nodes}
					explorerExpandedNodes={this.state.explorerExpandedNodes}
					view={this.state.detailsView}
					details={this.state.nodeDetails}
					topology={this.state.activeTopology} /> }

				<div className="navbar">
					<div className="navbar-brand">
						<div id="logo">
							<Logo />
						</div>
        		    </div>
					<Topologies topologies={this.state.topologies} active={this.state.activeTopology} />
					<Views nodes={this.state.nodes} active={this.state.currentView} />
					<ViewOptions active={this.state.activeTopologyMode}/>
				</div>

				{showingChord && <Chord nodes={this.state.nodes} />}
				{showingMatrix && <Matrix nodes={this.state.nodes} />}
				{showingNodes && <Nodes nodes={this.state.nodes} />}
				{showingAdjacency && <SearchBar nodes={this.state.nodes} filterText={this.state.filterText}
					filterAdjacent={this.state.filterAdjacent} />}
				{showingAdjacency && <Adjacency nodes={this.state.nodes} highlightedAdjacents={this.state.highlightedAdjacents}
					filterText={this.state.filterText} filterAdjacent={this.state.filterAdjacent} /> }
			</div>
		);
	}

});

module.exports = App;
