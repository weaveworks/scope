const React = require('react');
const mui = require('material-ui');
const Paper = mui.Paper;

const NodeDetails = require('./node-details');

const Details = React.createClass({

  render: function() {
    return (
      <div id="details">
        <Paper zDepth={3} style={{height: '100%', paddingBottom: 8}}>
          <NodeDetails {...this.props} />
        </Paper>
      </div>
    );
  }

});

module.exports = Details;
