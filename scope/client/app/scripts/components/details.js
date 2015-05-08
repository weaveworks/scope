/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');
var mui = require('material-ui');
var Paper = mui.Paper;
var IconButton = mui.IconButton;

var AppActions = require('../actions/app-actions');
var NodeDetails = require('./node-details');
var WebapiUtils = require('../utils/web-api-utils');

var Details = React.createClass({

	handleClickClose: function(ev) {
		ev.preventDefault();
		AppActions.clickCloseDetails();
	},

	render: function() {
		return (
			<div id="details">
				<Paper>
					<div className="details-tools">
						<span className="fa fa-close" onClick={this.handleClickClose} />
					</div>
					<NodeDetails details={this.props.details} />
				</Paper>
			</div>
		);
	}

});

module.exports = Details;