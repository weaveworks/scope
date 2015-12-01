import React from 'react';

import NodeDetailsTableRowValue from './node-details-table-row-value';
import NodeDetailsTableRowNumber from './node-details-table-row-number';
import NodeDetailsTableRowSparkline from './node-details-table-row-sparkline';

export default class NodeDetailsTable extends React.Component {
  render() {
    return (
      <div className="node-details-table">
        <h4 className="node-details-table-title truncate" title={this.props.title}>
          {this.props.title}
        </h4>

        {this.props.rows.map(function(row) {
          let valueComponent;
          if (row.value_type === 'numeric') {
            valueComponent = <NodeDetailsTableRowNumber row={row} />;
          } else if (row.value_type === 'sparkline') {
            valueComponent = <NodeDetailsTableRowSparkline row={row} />;
          } else {
            valueComponent = <NodeDetailsTableRowValue row={row} />;
          }
          return (
            <div className="node-details-table-row" key={row.key + row.value_major + row.value_minor}>
              <div className="node-details-table-row-key truncate" title={row.key}>{row.key}</div>
              {valueComponent}
            </div>
          );
        })}
      </div>
    );
  }
}
