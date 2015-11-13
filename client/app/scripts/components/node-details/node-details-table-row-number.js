const React = require('react');

const NodeDetailsTableRowNumber = React.createClass({

  render: function() {
    const row = this.props.row;
    return (
      <div className="node-details-table-row-value">
        <div className="node-details-table-row-value-scalar">{row.value_major}</div>
        <div className="node-details-table-row-value-unit">{row.value_minor}</div>
      </div>
    );
  }
});

module.exports = NodeDetailsTableRowNumber;
