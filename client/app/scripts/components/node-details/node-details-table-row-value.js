import React from 'react';

export default class NodeDetailsTableRowValue extends React.Component {
  render() {
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
}
