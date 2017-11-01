import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { find, get, union, sortBy, groupBy, concat, debounce } from 'lodash';

import { NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT } from '../../constants/limits';
import { TABLE_ROW_FOCUS_DEBOUNCE_INTERVAL } from '../../constants/timer';

import ShowMore from '../show-more';
import NodeDetailsTableRow from './node-details-table-row';
import NodeDetailsTableHeaders from './node-details-table-headers';
import { ipToPaddedString } from '../../utils/string-utils';
import { moveElement, insertElement } from '../../utils/array-utils';
import {
  isIP, isNumber, defaultSortDesc, getTableColumnsStyles
} from '../../utils/node-details-utils';


function getDefaultSortedBy(columns, nodes) {
  // default sorter specified by columns
  const defaultSortColumn = find(columns, {defaultSort: true});
  if (defaultSortColumn) {
    return defaultSortColumn.id;
  }
  // otherwise choose first metric
  const firstNodeWithMetrics = find(nodes, n => get(n, ['metrics', 0]));
  if (firstNodeWithMetrics) {
    return get(firstNodeWithMetrics, ['metrics', 0, 'id']);
  }

  return 'label';
}


function maybeToLower(value) {
  if (!value || !value.toLowerCase) {
    return value;
  }
  return value.toLowerCase();
}


function getNodeValue(node, header) {
  const fieldId = header && header.id;
  if (fieldId !== null) {
    let field = union(node.metrics, node.metadata).find(f => f.id === fieldId);

    if (field) {
      if (isIP(header)) {
        // Format the IPs so that they are sorted numerically.
        return ipToPaddedString(field.value);
      } else if (isNumber(header)) {
        return parseFloat(field.value);
      }
      return field.value;
    }

    if (node.parents) {
      field = node.parents.find(f => f.topologyId === fieldId);
      if (field) {
        return field.label;
      }
    }

    if (node[fieldId] !== undefined && node[fieldId] !== null) {
      return node[fieldId];
    }
  }

  return null;
}


function getValueForSortedBy(sortedByHeader) {
  return node => maybeToLower(getNodeValue(node, sortedByHeader));
}


function getMetaDataSorters(nodes) {
  // returns an array of sorters that will take a node
  return get(nodes, [0, 'metadata'], []).map((field, index) => (node) => {
    const nodeMetadataField = node.metadata && node.metadata[index];
    if (nodeMetadataField) {
      if (isNumber(nodeMetadataField)) {
        return parseFloat(nodeMetadataField.value);
      }
      return nodeMetadataField.value;
    }
    return null;
  });
}


function sortNodes(nodes, getValue, sortedDesc) {
  const sortedNodes = sortBy(
    nodes,
    getValue,
    getMetaDataSorters(nodes)
  );
  if (sortedDesc) {
    sortedNodes.reverse();
  }
  return sortedNodes;
}


function getSortedNodes(nodes, sortedByHeader, sortedDesc) {
  const getValue = getValueForSortedBy(sortedByHeader);
  const withAndWithoutValues = groupBy(nodes, (n) => {
    if (!n || n.valueEmpty) {
      return 'withoutValues';
    }
    const v = getValue(n);
    return v !== null && v !== undefined ? 'withValues' : 'withoutValues';
  });
  const withValues = sortNodes(withAndWithoutValues.withValues, getValue, sortedDesc);
  const withoutValues = sortNodes(withAndWithoutValues.withoutValues, getValue, sortedDesc);

  return concat(withValues, withoutValues);
}


// By inserting this fake invisible row into the table, with the help of
// some CSS trickery, we make the inner scrollable content of the table
// have a minimal height. That prevents auto-scroll under a focus if the
// number of table rows shrinks.
function minHeightConstraint(height = 0) {
  return <tr className="min-height-constraint" style={{height}} />;
}


class NodeDetailsTable extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      limit: props.limit,
      sortedDesc: this.props.sortedDesc,
      sortedBy: this.props.sortedBy
    };
    this.focusState = {};

    this.updateSorted = this.updateSorted.bind(this);
    this.handleLimitClick = this.handleLimitClick.bind(this);
    this.onMouseLeaveRow = this.onMouseLeaveRow.bind(this);
    this.onMouseEnterRow = this.onMouseEnterRow.bind(this);
    this.saveTableContentRef = this.saveTableContentRef.bind(this);
    // Use debouncing to prevent event flooding when e.g. crossing fast with mouse cursor
    // over the whole table. That would be expensive as each focus causes table to rerender.
    this.debouncedFocusRow = debounce(this.focusRow, TABLE_ROW_FOCUS_DEBOUNCE_INTERVAL);
    this.debouncedBlurRow = debounce(this.blurRow, TABLE_ROW_FOCUS_DEBOUNCE_INTERVAL);
  }

  updateSorted(sortedBy, sortedDesc) {
    this.setState({ sortedBy, sortedDesc });
    this.props.onSortChange(sortedBy, sortedDesc);
  }

  handleLimitClick() {
    const limit = this.state.limit ? 0 : this.props.limit;
    this.setState({ limit });
  }

  focusRow(rowIndex, node) {
    // Remember the focused row index, the node that was focused and
    // the table content height so that we can keep the node row fixed
    // without auto-scrolling happening.
    // NOTE: It would be ideal to modify the real component state here,
    // but that would cause whole table to rerender, which becomes to
    // expensive with the current implementation if the table consists
    // of 1000+ nodes.
    this.focusState = {
      focusedNode: node,
      focusedRowIndex: rowIndex,
      tableContentMinHeightConstraint: this.tableContent && this.tableContent.scrollHeight,
    };
  }

  blurRow() {
    // Reset the focus state
    this.focusState = {};
  }

  onMouseEnterRow(rowIndex, node) {
    this.debouncedBlurRow.cancel();
    this.debouncedFocusRow(rowIndex, node);
  }

  onMouseLeaveRow() {
    this.debouncedFocusRow.cancel();
    this.debouncedBlurRow();
  }

  saveTableContentRef(ref) {
    this.tableContent = ref;
  }

  getColumnHeaders() {
    const columns = this.props.columns || [];
    return [{id: 'label', label: this.props.label}].concat(columns);
  }

  render() {
    const {
      nodeIdKey, columns, topologyId, onClickRow,
      onMouseEnter, onMouseLeave, timestamp
    } = this.props;

    const sortedBy = this.state.sortedBy || getDefaultSortedBy(columns, this.props.nodes);
    const sortedByHeader = this.getColumnHeaders().find(h => h.id === sortedBy);
    const sortedDesc = (this.state.sortedDesc === null) ?
      defaultSortDesc(sortedByHeader) : this.state.sortedDesc;

    let nodes = getSortedNodes(this.props.nodes, sortedByHeader, sortedDesc);

    const { focusedNode, focusedRowIndex, tableContentMinHeightConstraint } = this.focusState;
    if (Number.isInteger(focusedRowIndex) && focusedRowIndex < nodes.length) {
      const nodeRowIndex = nodes.findIndex(node => node.id === focusedNode.id);
      if (nodeRowIndex >= 0) {
        // If the focused node still exists in the table, we move it
        // to the hovered row, keeping the rest of the table sorted.
        nodes = moveElement(nodes, nodeRowIndex, focusedRowIndex);
      } else {
        // Otherwise we insert the dead focused node there, pretending
        // it's still alive. That enables the users to read off all the
        // info they want and perhaps even open the details panel. Also,
        // only if we do this, we can guarantee that mouse hover will
        // always freeze the table row until we focus out.
        nodes = insertElement(nodes, focusedRowIndex, focusedNode);
      }
    }

    // If we are 1 over the limit, we still show the row. We never display
    // "+1" but only "+2" and up.
    const limit = this.state.limit > 0 && nodes.length === this.state.limit + 1
      ? nodes.length
      : this.state.limit;
    const limited = nodes && limit > 0 && nodes.length > limit;
    const expanded = limit === 0;
    const notShown = nodes.length - limit;
    if (nodes && limited) {
      nodes = nodes.slice(0, limit);
    }

    const className = classNames('node-details-table-wrapper-wrapper', this.props.className);
    const headers = this.getColumnHeaders();
    const styles = getTableColumnsStyles(headers);

    return (
      <div className={className} style={this.props.style}>
        <div className="node-details-table-wrapper">
          <table className="node-details-table">
            <thead>
              {this.props.nodes && this.props.nodes.length > 0 && <NodeDetailsTableHeaders
                headers={headers}
                sortedBy={sortedBy}
                sortedDesc={sortedDesc}
                onClick={this.updateSorted}
              />}
            </thead>
            <tbody
              style={this.props.tbodyStyle}
              ref={this.saveTableContentRef}
              onMouseEnter={onMouseEnter}
              onMouseLeave={onMouseLeave}>
              {nodes && nodes.map((node, index) => (
                <NodeDetailsTableRow
                  key={node.id}
                  renderIdCell={this.props.renderIdCell}
                  selected={this.props.selectedNodeId === node.id}
                  node={node}
                  index={index}
                  nodeIdKey={nodeIdKey}
                  colStyles={styles}
                  columns={columns}
                  onClick={onClickRow}
                  onMouseEnter={this.onMouseEnterRow}
                  onMouseLeave={this.onMouseLeaveRow}
                  timestamp={timestamp}
                  topologyId={topologyId} />
              ))}
              {minHeightConstraint(tableContentMinHeightConstraint)}
            </tbody>
          </table>
          <ShowMore
            handleClick={this.handleLimitClick}
            collection={nodes}
            expanded={expanded}
            notShown={notShown} />
        </div>
      </div>
    );
  }
}


NodeDetailsTable.defaultProps = {
  nodeIdKey: 'id', // key to identify a node in a row (used for topology links)
  limit: NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT,
  onSortChange: () => {},
  sortedDesc: null,
  sortedBy: null,
};

function mapStateToProps(state) {
  return {
    timestamp: state.get('pausedAt'),
  };
}

export default connect(mapStateToProps)(NodeDetailsTable);
