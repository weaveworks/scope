import _ from 'lodash';
import React from 'react';
import classNames from 'classnames';

import ShowMore from '../show-more';
import NodeDetailsTableRow from './node-details-table-row';


function isNumberField(field) {
  return field.dataType && field.dataType === 'number';
}

const CW = {
  S: '50px',
  M: '80px',
  L: '120px',
  XL: '140px',
  XXL: '170px',
};


const COLUMN_WIDTHS = {
  count: '60px',
  docker_container_created: CW.XL,
  docker_container_restart_count: CW.M,
  docker_container_state_human: CW.XXL,
  docker_container_uptime: '85px',
  docker_cpu_total_usage: CW.M,
  docker_memory_usage: CW.M,
  remote: CW.L,
  open_files_count: CW.M,
  pid: CW.M,
  port: CW.S,
  ppid: CW.M,
  process_cpu_usage_percent: CW.M,
  process_memory_usage_bytes: CW.M,
  threads: CW.M,
  // e.g. details panel > pods
  kubernetes_ip: CW.L,
};


function getDefaultSortBy(columns, nodes) {
  // default sorter specified by columns
  const defaultSortColumn = _.find(columns, {defaultSort: true});
  if (defaultSortColumn) {
    return defaultSortColumn.id;
  }
  // otherwise choose first metric
  const firstNodeWithMetrics = _.find(nodes, n => _.get(n, ['metrics', 0]));
  if (firstNodeWithMetrics) {
    return _.get(firstNodeWithMetrics, ['metrics', 0, 'id']);
  }

  return 'label';
}


function maybeToLower(value) {
  if (!value || !value.toLowerCase) {
    return value;
  }
  return value.toLowerCase();
}


function getNodeValue(node, fieldId) {
  if (fieldId !== null) {
    let field = _.union(node.metrics, node.metadata).find(f => f.id === fieldId);

    if (field) {
      if (isNumberField(field)) {
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


function getValueForSortBy(sortBy) {
  return (node) => maybeToLower(getNodeValue(node, sortBy));
}


function getMetaDataSorters(nodes) {
  // returns an array of sorters that will take a node
  return _.get(nodes, [0, 'metadata'], []).map((field, index) => node => {
    const nodeMetadataField = node.metadata && node.metadata[index];
    if (nodeMetadataField) {
      if (isNumberField(nodeMetadataField)) {
        return parseFloat(nodeMetadataField.value);
      }
      return nodeMetadataField.value;
    }
    return null;
  });
}


function sortNodes(nodes, getValue, sortedDesc) {
  const sortedNodes = _.sortBy(
    nodes,
    getValue,
    getMetaDataSorters(nodes)
  );
  if (sortedDesc) {
    sortedNodes.reverse();
  }
  return sortedNodes;
}


function getSortedNodes(nodes, sortBy, sortedDesc) {
  const getValue = getValueForSortBy(sortBy);
  const withAndWithoutValues = _.groupBy(nodes, (n) => {
    const v = getValue(n);
    return v !== null && v !== undefined ? 'withValues' : 'withoutValues';
  });
  const withValues = sortNodes(withAndWithoutValues.withValues, getValue, sortedDesc);
  const withoutValues = sortNodes(withAndWithoutValues.withoutValues, getValue, sortedDesc);

  return _.concat(withValues, withoutValues);
}


function getColumnWidth(headers, h, i) {
  //
  // Beauty hack: adjust first column width if there are only few columns;
  // this assumes the other columns are narrow metric columns of 20% table width
  //
  if (i === 0) {
    if (headers.length === 2) {
      return '66%';
    } else if (headers.length === 3) {
      return '50%';
    } else if (headers.length > 3 && headers.length <= 5) {
      return '33%';
    }
  }

  //
  // More beauty hacking, ports and counts can only get so big, free up WS for other longer
  // fields like IPs!
  //
  return COLUMN_WIDTHS[h.id];
}


function getColumnsStyles(headers) {
  return headers.map((h, i) => ({
    width: getColumnWidth(headers, h, i),
    textAlign: h.dataType === 'number' ? 'right' : 'left',
  }));
}


function defaultSortDesc(header) {
  return header.dataType === 'number';
}


export default class NodeDetailsTable extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.DEFAULT_LIMIT = 5;
    this.state = {
      limit: props.limit || this.DEFAULT_LIMIT,
      sortedDesc: this.props.sortedDesc,
      sortBy: this.props.sortBy
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
  }

  handleHeaderClick(ev, headerId, currentSortBy, currentSortedDesc) {
    ev.preventDefault();
    const header = this.getColumnHeaders().find(h => h.id === headerId);
    const sortBy = header.id;
    const sortedDesc = header.id === currentSortBy
      ? !currentSortedDesc : defaultSortDesc(header);
    this.setState({sortBy, sortedDesc});
    this.props.onSortChange(sortBy, sortedDesc);
  }

  handleLimitClick() {
    const limit = this.state.limit ? 0 : this.DEFAULT_LIMIT;
    this.setState({limit});
  }

  getColumnHeaders() {
    const columns = this.props.columns || [];
    return [{id: 'label', label: this.props.label}].concat(columns);
  }

  renderHeaders(sortBy, sortedDesc) {
    if (this.props.nodes && this.props.nodes.length > 0) {
      const headers = this.getColumnHeaders();
      const colStyles = getColumnsStyles(headers);

      return (
        <tr>
          {headers.map((header, i) => {
            const headerClasses = ['node-details-table-header', 'truncate'];
            const onHeaderClick = ev => {
              this.handleHeaderClick(ev, header.id, sortBy, sortedDesc);
            };
            // sort by first metric by default
            const isSorted = header.id === sortBy;
            const isSortedDesc = isSorted && sortedDesc;
            const isSortedAsc = isSorted && !isSortedDesc;

            if (isSorted) {
              headerClasses.push('node-details-table-header-sorted');
            }

            return (
              <td className={headerClasses.join(' ')} style={colStyles[i]} onClick={onHeaderClick}
                title={header.label} key={header.id}>
                {isSortedAsc
                  && <span className="node-details-table-header-sorter fa fa-caret-up" />}
                {isSortedDesc
                  && <span className="node-details-table-header-sorter fa fa-caret-down" />}
                {header.label}
              </td>
            );
          })}
        </tr>
      );
    }
    return '';
  }

  render() {
    const { nodeIdKey, columns, topologyId, onClickRow, onMouseEnter, onMouseLeave,
      onMouseEnterRow, onMouseLeaveRow } = this.props;

    const sortBy = this.state.sortBy || getDefaultSortBy(columns, this.props.nodes);
    const header = this.getColumnHeaders().find(h => h.id === sortBy);
    const sortedDesc = this.state.sortedDesc !== null ?
      this.state.sortedDesc :
      defaultSortDesc(header);

    let nodes = getSortedNodes(this.props.nodes, sortBy, sortedDesc);
    const limited = nodes && this.state.limit > 0 && nodes.length > this.state.limit;
    const expanded = this.state.limit === 0;
    const notShown = nodes.length - this.state.limit;
    if (nodes && limited) {
      nodes = nodes.slice(0, this.state.limit);
    }

    const className = classNames('node-details-table-wrapper-wrapper', this.props.className);

    return (
      <div className={className}
        style={this.props.style}>
        <div className="node-details-table-wrapper">
          <table className="node-details-table">
            <thead>
              {this.renderHeaders(sortBy, sortedDesc)}
            </thead>
            <tbody style={this.props.tbodyStyle} onMouseEnter={onMouseEnter}
              onMouseLeave={onMouseLeave}>
              {nodes && nodes.map(node => (
                <NodeDetailsTableRow
                  key={node.id}
                  renderIdCell={this.props.renderIdCell}
                  selected={this.props.selectedNodeId === node.id}
                  node={node}
                  nodeIdKey={nodeIdKey}
                  colStyles={getColumnsStyles(this.getColumnHeaders())}
                  columns={columns}
                  onClick={onClickRow}
                  onMouseLeaveRow={onMouseLeaveRow}
                  onMouseEnterRow={onMouseEnterRow}
                  topologyId={topologyId} />
              ))}
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
  nodeIdKey: 'id',  // key to identify a node in a row (used for topology links)
  onSortChange: () => {},
};
