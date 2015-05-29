const React = require('react');

const NodeDetailsTable = React.createClass({

  render: function() {
    const isNumeric = this.props.isNumeric;

    return (
      <div className="node-details-table">
        <h4 className="node-details-table-title">
          {this.props.title}
        </h4>

        {this.props.rows.map(function(row) {
          return (
            <div className="node-details-table-row">
              <div className="node-details-table-row-key">{row.key}</div>
              {isNumeric && <div className="node-details-table-row-value-scalar">{row.value_major}</div>}
              {isNumeric && <div className="node-details-table-row-value-unit">{row.value_minor}</div>}
              {!isNumeric && <div className="node-details-table-row-value-major truncate">{row.value_major}</div>}
              {!isNumeric && row.value_minor && <div className="node-details-table-row-value-minor truncate">{row.value_minor}</div>}
            </div>
          );
        })}
      </div>
    );
  }

});

module.exports = NodeDetailsTable;
