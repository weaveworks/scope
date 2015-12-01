import React from 'react';

export default class NodeDetailsTableRowNumber extends React.Component {
  render() {
    const row = this.props.row;
    return (
      <div className="node-details-table-row-value">
        <div className="node-details-table-row-value-scalar">{row.value_major}</div>
        <div className="node-details-table-row-value-unit">{row.value_minor}</div>
      </div>
    );
  }
}
