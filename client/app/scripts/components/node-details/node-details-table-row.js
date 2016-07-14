import React from 'react';
import classNames from 'classnames';

import NodeDetailsTableNodeLink from './node-details-table-node-link';
import NodeDetailsTableNodeMetric from './node-details-table-node-metric';


function getValuesForNode(node) {
  const values = {};
  ['metrics', 'metadata'].forEach(collection => {
    if (node[collection]) {
      node[collection].forEach(field => {
        const result = Object.assign({}, field);
        result.valueType = collection;
        values[field.id] = result;
      });
    }
  });

  (node.parents || []).forEach(p => {
    values[p.topologyId] = {
      id: p.topologyId,
      label: p.topologyId,
      value: p.label,
      relative: p,
      valueType: 'relatives',
    };
  });

  return values;
}


function renderValues(node, columns = [], columnWidths = []) {
  const fields = getValuesForNode(node);
  return columns.map(({id}, i) => {
    const field = fields[id];
    const style = { width: columnWidths[i] };
    if (field) {
      if (field.valueType === 'metadata') {
        return (
          <td className="node-details-table-node-value truncate" title={field.value}
            style={style}
            key={field.id}>
            {field.value}
          </td>
        );
      }
      if (field.valueType === 'relatives') {
        return (
          <td className="node-details-table-node-value truncate" title={field.value}
            style={style}
            key={field.id}>
            {<NodeDetailsTableNodeLink linkable nodeId={field.relative.id} {...field.relative} />}
          </td>
        );
      }
      return <NodeDetailsTableNodeMetric style={style} key={field.id} {...field} />;
    }
    // empty cell to complete the row for proper hover
    return <td className="node-details-table-node-value" style={style} key={id} />;
  });
}


export default class NodeDetailsTableRow extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.onMouseEnter = this.onMouseEnter.bind(this);
    this.onMouseLeave = this.onMouseLeave.bind(this);
  }

  onMouseEnter() {
    const { node, onMouseEnterRow } = this.props;
    onMouseEnterRow(node);
  }

  onMouseLeave() {
    const { node, onMouseLeaveRow } = this.props;
    onMouseLeaveRow(node);
  }

  render() {
    const { node, nodeIdKey, topologyId, columns, onMouseEnterRow, onMouseLeaveRow, selected,
      widths } = this.props;
    const [firstColumnWidth, ...columnWidths] = widths;
    const values = renderValues(node, columns, columnWidths);
    const nodeId = node[nodeIdKey];
    const className = classNames('node-details-table-node', { selected });

    return (
      <tr
        onMouseEnter={onMouseEnterRow && this.onMouseEnter}
        onMouseLeave={onMouseLeaveRow && this.onMouseLeave}
        className={className}>
        <td className="node-details-table-node-label truncate"
          style={{ width: firstColumnWidth }}>
          {this.props.renderIdCell(Object.assign(node, {topologyId, nodeId}))}
        </td>
        {values}
      </tr>
    );
  }
}


NodeDetailsTableRow.defaultProps = {
  renderIdCell: (props) => <NodeDetailsTableNodeLink {...props} />
};
