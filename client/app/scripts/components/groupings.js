/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');
var AppStore = require('../stores/app-store');

var GROUPINGS = [{
	id: 'none',
	iconClass: 'fa fa-th'
}, {
	id: 'grouped',
	iconClass: 'fa fa-th-large'
}];

var Groupings = React.createClass({

	onGroupingClick: function(ev) {
		ev.preventDefault();
		AppActions.clickGrouping(ev.currentTarget.getAttribute('rel'));
	},

	renderGrouping: function(grouping, active) {
		var className = grouping.id === active ? "groupings-item groupings-item-active" : "groupings-item";

		return (
			<div className={className} key={grouping.id} rel={grouping.id} onClick={this.onGroupingClick}>
				<span className={grouping.iconClass} />
			</div>
		);
	},

	render: function() {
		var activeGrouping = this.props.active;

		return (
			<div className="groupings">
				{GROUPINGS.map(function(grouping) {
					return this.renderGrouping(grouping, activeGrouping);
				}, this)}
			</div>
		);
	}

});

module.exports = Groupings;
