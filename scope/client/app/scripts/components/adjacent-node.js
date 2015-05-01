/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');
var TopologyActions = require('../actions/topology-actions');

var AdjacentNode = React.createClass({

	_enterTriggered: false,

	onClick: function(nodeId, ev) {
		AppActions.clickNode(nodeId);
	},

	onMouseEnter: function(nodeId, ev) {
		var cmp = this;

		// delay enter node action

		this._mouseTransaction = _.delay(function(nodeId) {
			cmp._enterTriggered = nodeId;
			TopologyActions.enterNode(nodeId);
		}, 1000, nodeId);
	},

	onMouseLeave: function(nodeId, ev) {
		clearTimeout(this._mouseTransaction);
		if (this._enterTriggered === nodeId) {
			TopologyActions.leaveNode(nodeId);
		}
	},

	highlightAdjacent: function(highlightedAdjacents, node) {
		return highlightedAdjacents && !_.contains(highlightedAdjacents, node.id) ? 'non-adjacent' : '';
	},

	highlightFilter: function(label, filterText, allowFilter) {
		var startIndex = label.toLowerCase().indexOf(filterText.toLowerCase()),
			stopIndex = startIndex + filterText.length,
			pre, match, post;

		if (allowFilter && filterText && startIndex > -1) {
			return (
				<span>
					<span>{label.substring(0, startIndex)}</span>
					<span className="highlight-filter">{label.substring(startIndex, stopIndex)}</span>
					<span>{label.substring(stopIndex)}</span>
				</span>
			);
		}

		return label;
	},

	render: function() {
		var filterText = this.props.filterText,
			allowFilter = this.props.allowFilter,
			classNames = 'btn btn-default node ' + this.highlightAdjacent(this.props.highlightedAdjacents, this.props.node),
			major = this.highlightFilter(this.props.node.label_major, filterText, allowFilter),
			minor = this.props.node.label_minor && this.highlightFilter(this.props.node.label_minor, filterText, allowFilter);

		return (
			<a className={classNames}
				onClick={this.onClick.bind(this, this.props.node.id)} 
				onMouseEnter={this.onMouseEnter.bind(this, this.props.node.id)} 
				onMouseLeave={this.onMouseLeave.bind(this, this.props.node.id)}>
				<span className="major">{major}</span>
				<span className="minor">{minor}</span>
			</a>
		);
	}

});

module.exports = AdjacentNode;