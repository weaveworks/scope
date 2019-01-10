import React from 'react';
import classNames from 'classnames';
import { groupBy, mapValues } from 'lodash';
import { intersperse } from '../../utils/array-utils';


import NodeDetailsTableNodeLink from './node-details-table-node-link';
import NodeDetailsTableNodeMetricLink from './node-details-table-node-metric-link';
import { formatDataType } from '../../utils/string-utils';

function getValuesForNode(node) {
  let values = {};
  ['metrics', 'metadata'].forEach((collection) => {
    if (node[collection]) {
      node[collection].forEach((field) => {
        const result = Object.assign({}, field);
        result.valueType = collection;
        values[field.id] = result;
      });
    }
  });

  if (node.parents) {
    const byTopologyId = groupBy(node.parents, parent => parent.topologyId);
    const relativesByTopologyId = mapValues(byTopologyId, (relatives, topologyId) => ({
      id: topologyId,
      label: topologyId,
      relatives,
      value: relatives.map(relative => relative.label).join(', '),
      valueType: 'relatives',
    }));

    values = {
      ...values,
      ...relativesByTopologyId,
    };
  }

  return values;
}


function renderValues(node, columns = [], columnStyles = [], timestamp = null, topologyId = null) {
  const fields = getValuesForNode(node);
  return columns.map(({ id }, i) => {
    const field = fields[id];
    const style = columnStyles[i];
    if (field) {
      if (field.valueType === 'metadata') {
        const { value, title } = formatDataType(field, timestamp);
        return (
          <td
            className="node-details-table-node-value truncate"
            title={title}
            style={style}
            key={field.id}>
            {field.dataType === 'link' ?
              <a
                rel="noopener noreferrer"
                target="_blank"
                className="node-details-table-node-link"
                href={value}>{value}
              </a> :
              value}
          </td>
        );
      }
      if (field.valueType === 'relatives') {
        return (
          <td
            className="node-details-table-node-value truncate"
            title={field.value}
            style={style}
            key={field.id}>
            {intersperse(field.relatives.map(relative =>
              (<NodeDetailsTableNodeLink
                key={relative.id}
                linkable
                nodeId={relative.id}
                {...relative}
              />)), ' ')}
          </td>
        );
      }
      // valueType === 'metrics'
      return (
        <NodeDetailsTableNodeMetricLink
          style={style}
          key={field.id}
          topologyId={topologyId}
          {...field} />
      );
    }
    // empty cell to complete the row for proper hover
    return (
      <td className="node-details-table-node-value" style={style} key={id} />
    );
  });
}

/**
 * Table row children may react to onClick events but the row
 * itself does detect a click by looking at onMouseUp. To stop
 * the bubbling of clicks on child elements we need to dismiss
 * the onMouseUp event.
 */
export const dismissRowClickProps = {
  onMouseUp: (ev) => {
    ev.preventDefault();
    ev.stopPropagation();
  }
};

export default class NodeDetailsTableRow extends React.Component {
  constructor(props, context) {
    super(props, context);

    //
    // We watch how far the mouse moves when click on a row, move to much and we assume that the
    // user is selecting some data in the row. In this case don't trigger the onClick event which
    // is most likely a details panel popping open.
    //
    this.state = { focused: false };
    this.mouseDrag = {};
  }

  onMouseEnter = () => {
    this.setState({ focused: true });
    if (this.props.onMouseEnter) {
      this.props.onMouseEnter(this.props.index, this.props.node);
    }
  }

  onMouseLeave = () => {
    this.setState({ focused: false });
    if (this.props.onMouseLeave) {
      this.props.onMouseLeave();
    }
  }

  onMouseDown = (ev) => {
    this.mouseDrag = {
      originX: ev.pageX,
      originY: ev.pageY,
    };
  }

  onClick = (ev) => {
    const thresholdPx = 2;
    const { pageX, pageY } = ev;
    const { originX, originY } = this.mouseDrag;
    const movedTheMouseTooMuch = (
      Math.abs(originX - pageX) > thresholdPx ||
      Math.abs(originY - pageY) > thresholdPx
    );
    if (movedTheMouseTooMuch && originX && originY) {
      return;
    }

    this.props.onClick(ev, this.props.node);
    this.mouseDrag = {};
  }

  render() {
    const {
      node, nodeIdKey, topologyId, columns, onClick, colStyles, timestamp
    } = this.props;
    const [firstColumnStyle, ...columnStyles] = colStyles;
    const values = renderValues(node, columns, columnStyles, timestamp, topologyId);
    const nodeId = node[nodeIdKey];

    const className = classNames('tour-step-anchor node-details-table-node', {
      focused: this.state.focused,
      selected: this.props.selected,
    });

    return (
      <tr
        onClick={onClick && this.onClick}
        onMouseDown={onClick && this.onMouseDown}
        onMouseEnter={this.onMouseEnter}
        onMouseLeave={this.onMouseLeave}
        className={className}>
        <td className="node-details-table-node-label truncate" style={firstColumnStyle}>
          {this.props.renderIdCell(Object.assign(node, {nodeId, topologyId}))}
        </td>
        {values}
      </tr>
    );
  }
}


NodeDetailsTableRow.defaultProps = {
  renderIdCell: props => <NodeDetailsTableNodeLink {...props} />
};
