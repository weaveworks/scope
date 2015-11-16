const React = require('react');

const NodeDetailsTableRowValue = React.createClass({

  render: function() {
    const row = this.props.row;
    return (
      <div className="node-details-table-row-value">
        <div className="node-details-table-row-value-major truncate" title={row.value_major}>
          {row.value_major}
        </div>
        {row.value_minor && <div className="node-details-table-row-value-minor truncate" title={row.value_minor}>
          {row.value_minor}
        </div>}
      </div>
    );
  }
});

module.exports = NodeDetailsTableRowValue;
