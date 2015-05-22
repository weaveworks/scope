/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');
var AppStore = require('../stores/app-store');

var GROUPINGS = [{
	id: 'none',
	iconClass: 'fa fa-th',
	needsTopology: false
}, {
	id: 'grouped',
	iconClass: 'fa fa-th-large',
	needsTopology: 'grouped_url'
}];

var Groupings = React.createClass({

	onGroupingClick: function(ev) {
		ev.preventDefault();
		AppActions.clickGrouping(ev.currentTarget.getAttribute('rel'));
	},

	isGroupingSupportedByTopology: function(topology, grouping) {
		return !grouping.needsTopology || topology && topology[grouping.needsTopology];
	},

	getGroupingsSupportedByTopology: function(topology) {
		return _.filter(GROUPINGS, _.partial(this.isGroupingSupportedByTopology, topology));
	},

	renderGrouping: function(grouping, activeGroupingId) {
		var className = "groupings-item",
			isSupportedByTopology = this.isGroupingSupportedByTopology(this.props.currentTopology, grouping);

		if (grouping.id === activeGroupingId) {
			className += " groupings-item-active";
		} else if (!isSupportedByTopology) {
			className += " groupings-item-disabled";
		} else {
			className += " groupings-item-default";
		}

		return (
			<div className={className} key={grouping.id} rel={grouping.id} onClick={isSupportedByTopology && this.onGroupingClick}>
				<span className={grouping.iconClass} />
			</div>
		);
	},

	render: function() {
		var activeGrouping = this.props.active,
			isGroupingSupported = _.size(this.getGroupingsSupportedByTopology(this.props.currentTopology)) > 1;

		return (
			<div className="groupings">
				{isGroupingSupported && GROUPINGS.map(function(grouping) {
					return this.renderGrouping(grouping, activeGrouping);
				}, this)}
			</div>
		);
	}

});

module.exports = Groupings;
