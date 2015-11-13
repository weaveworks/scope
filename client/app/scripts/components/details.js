const React = require('react');

const AppActions = require('../actions/app-actions');
const NodeDetails = require('./node-details');

const Details = React.createClass({

  handleClickClose: function(ev) {
    ev.preventDefault();
    AppActions.clickCloseDetails();
  },

  render: function() {
    return (
      <div id="details">
        <div style={{height: '100%', paddingBottom: 8, borderRadius: 2,
          backgroundColor: '#fff',
          boxShadow: '0 10px 30px rgba(0, 0, 0, 0.19), 0 6px 10px rgba(0, 0, 0, 0.23)'}}>
          <div className="details-tools-wrapper">
            <div className="details-tools">
              <span className="fa fa-close" onClick={this.handleClickClose} />
            </div>
          </div>
          <NodeDetails {...this.props} />
        </div>
      </div>
    );
  }

});

module.exports = Details;
