const React = require('react');
const mui = require('material-ui');
const Paper = mui.Paper;

const AppActions = require('../actions/app-actions');
const NodeDetails = require('./node-details');

const Details = React.createClass({

  render: function() {
    return (
      <div id="details">
        <Paper zDepth={3}>
          <div className="details-tools-wrapper">
            <div className="details-tools">
              <span className="fa fa-close" onClick={this.handleClickClose} />
            </div>
          </div>
          <NodeDetails nodeId={this.props.nodeId} details={this.props.details}
            nodes={this.props.nodes} />
        </Paper>
      </div>
    );
  },

  handleClickClose: function(ev) {
    ev.preventDefault();
    AppActions.clickCloseDetails();
  }

});

module.exports = Details;
