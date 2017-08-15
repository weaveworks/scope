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
      value: relatives.map(relative => relative.label).join(', '),
      valueType: 'relatives',
      relatives,
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
            {value}
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
              <NodeDetailsTableNodeLink
                key={relative.id}
                linkable
                nodeId={relative.id}
                {...relative}
              />
            ), ' ')}
          </td>
        );
      }
      // valueType === 'metrics'
      return (
        <NodeDetailsTableNodeMetricLink
          style={style} key={field.id} topologyId={topologyId} {...field} />
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
    this.mouseDragOrigin = [0, 0];

    this.onMouseDown = this.onMouseDown.bind(this);
    this.onMouseUp = this.onMouseUp.bind(this);
    this.onMouseEnter = this.onMouseEnter.bind(this);
    this.onMouseLeave = this.onMouseLeave.bind(this);
  }

  onMouseEnter() {
    this.setState({ focused: true });
    if (this.props.onMouseEnter) {
      this.props.onMouseEnter(this.props.index, this.props.node);
    }
  }

  onMouseLeave() {
    this.setState({ focused: false });
    if (this.props.onMouseLeave) {
      this.props.onMouseLeave();
    }
  }

  onMouseDown(ev) {
    const { pageX, pageY } = ev;
    this.mouseDragOrigin = [pageX, pageY];
  }

  onMouseUp(ev) {
    const [originX, originY] = this.mouseDragOrigin;
    const { pageX, pageY } = ev;
    const thresholdPx = 2;
    const movedTheMouseTooMuch = (
      Math.abs(originX - pageX) > thresholdPx ||
      Math.abs(originY - pageY) > thresholdPx
    );
    if (movedTheMouseTooMuch) {
      return;
    }

    this.props.onClick(ev, this.props.node);
  }

  render() {
    const { node, nodeIdKey, topologyId, columns, onClick, colStyles, timestamp } = this.props;
    const [firstColumnStyle, ...columnStyles] = colStyles;
    const values = renderValues(node, columns, columnStyles, timestamp, topologyId);
    const nodeId = node[nodeIdKey];

    const className = classNames('node-details-table-node', {
      selected: this.props.selected,
      focused: this.state.focused,
    });

    return (
      <tr
        onMouseDown={onClick && this.onMouseDown}
        onMouseUp={onClick && this.onMouseUp}
        onMouseEnter={this.onMouseEnter}
        onMouseLeave={this.onMouseLeave}
        className={className}>
        <td className="node-details-table-node-label truncate" style={firstColumnStyle}>
          {this.props.renderIdCell(Object.assign(node, {topologyId, nodeId}))}
        </td>
        {values}
      </tr>
    );
  }
}


NodeDetailsTableRow.defaultProps = {
  renderIdCell: props => <NodeDetailsTableNodeLink {...props} />
};
