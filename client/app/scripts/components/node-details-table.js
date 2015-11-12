const React = require('react');
const Sparkline = require('./sparkline');

const NodeDetailsTable = React.createClass({

  render: function() {
    return (
      <div className="node-details-table">
        <h4 className="node-details-table-title truncate" title={this.props.title}>
          {this.props.title}
        </h4>

        {this.props.rows.map(function(row) {
          return (
            <div className="node-details-table-row" key={row.key + row.value_major}>
              <div className="node-details-table-row-key truncate" title={row.key}>{row.key}</div>
              { row.value_type === 'numeric' && <div className="node-details-table-row-value-scalar">{row.value_major}</div> }
              { row.value_type === 'numeric' && <div className="node-details-table-row-value-unit">{row.value_minor}</div> }
              { row.value_type === 'sparkline' && <div className="node-details-table-row-value-sparkline"><Sparkline data={row.metric.samples} min={0} max={row.metric.max} first={row.metric.first} last={row.metric.last} interpolate="none" />{row.value_major}</div> }
              { row.value_type === 'sparkline' && <div className="node-details-table-row-value-unit">{row.value_minor}</div> }
              { row.value_type !== 'numeric' && row.value_type !== 'sparkline' && <div className="node-details-table-row-value-major truncate" title={row.value_major}>{row.value_major}</div> }
              { row.value_type !== 'numeric' && row.value_type !== 'sparkline' && row.value_minor && <div className="node-details-table-row-value-minor truncate" title={row.value_minor}>{row.value_minor}</div> }
            </div>
          );
        })}
      </div>
    );
  }

});

module.exports = NodeDetailsTable;
