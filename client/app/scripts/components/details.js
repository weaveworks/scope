const React = require('react');

const NodeDetails = require('./node-details');

const Details = React.createClass({

  render: function() {
    return (
      <div id="details">
        <div style={{height: '100%', paddingBottom: 8, borderRadius: 2,
          backgroundColor: '#fff',
          boxShadow: '0 10px 30px rgba(0, 0, 0, 0.19), 0 6px 10px rgba(0, 0, 0, 0.23)'}}>
          <NodeDetails {...this.props} />
        </div>
      </div>
    );
  }

});

module.exports = Details;
