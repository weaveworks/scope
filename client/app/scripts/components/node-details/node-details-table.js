import _ from 'lodash';
import React from 'react';
import classNames from 'classnames';

import ShowMore from '../show-more';
import NodeDetailsTableRow from './node-details-table-row';


function isNumber(data) {
  return data.dataType && data.dataType === 'number';
}

const CW = {
  XS: '32px',
  S: '50px',
  M: '70px',
  L: '120px',
  XL: '140px',
  XXL: '170px',
};


const XS_LABEL = {
  count: '#',
  // TODO: consider changing the name of this field on the BE
  container: '#',
};


const COLUMN_WIDTHS = {
  count: CW.XS,
  container: CW.XS,
  docker_container_created: CW.XXL,
  docker_container_restart_count: CW.M,
  docker_container_state_human: CW.XXL,
  docker_container_uptime: '85px',
  docker_cpu_total_usage: CW.M,
  docker_memory_usage: CW.M,
  open_files_count: CW.M,
  pid: CW.S,
  port: CW.S,
  ppid: CW.S,
  process_cpu_usage_percent: CW.M,
  process_memory_usage_bytes: CW.M,
  threads: CW.M,

  // e.g. details panel > pods
  kubernetes_ip: CW.L,
  kubernetes_state: CW.M,
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


function getNodeValue(node, header) {
  const fieldId = header && header.id;
  if (fieldId !== null) {
    let field = _.union(node.metrics, node.metadata).find(f => f.id === fieldId);

    if (field) {
      if (isNumber(header)) {
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


function getValueForSortBy(sortByHeader) {
  return (node) => maybeToLower(getNodeValue(node, sortByHeader));
}


function getMetaDataSorters(nodes) {
  // returns an array of sorters that will take a node
  return _.get(nodes, [0, 'metadata'], []).map((field, index) => node => {
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


function getSortedNodes(nodes, sortByHeader, sortedDesc) {
  const getValue = getValueForSortBy(sortByHeader);
  const withAndWithoutValues = _.groupBy(nodes, (n) => {
    const v = getValue(n);
    return v !== null && v !== undefined ? 'withValues' : 'withoutValues';
  });
  const withValues = sortNodes(withAndWithoutValues.withValues, getValue, sortedDesc);
  const withoutValues = sortNodes(withAndWithoutValues.withoutValues, getValue, sortedDesc);

  return _.concat(withValues, withoutValues);
}


function getColumnWidth(headers, h) {
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
  return header && header.dataType === 'number';
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
    if (!this.props.nodes || this.props.nodes.length === 0) {
      return null;
    }

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

          const style = colStyles[i];
          const label = (style.width === CW.XS && XS_LABEL[header.id]) ?
            XS_LABEL[header.id] :
            header.label;

          return (
            <td className={headerClasses.join(' ')} style={style} onClick={onHeaderClick}
              title={header.label} key={header.id}>
              {isSortedAsc
                && <span className="node-details-table-header-sorter fa fa-caret-up" />}
              {isSortedDesc
                && <span className="node-details-table-header-sorter fa fa-caret-down" />}
              {label}
            </td>
          );
        })}
      </tr>
    );
  }

  render() {
    const { nodeIdKey, columns, topologyId, onClickRow, onMouseEnter, onMouseLeave,
      onMouseEnterRow, onMouseLeaveRow } = this.props;

    const sortBy = this.state.sortBy || getDefaultSortBy(columns, this.props.nodes);
    const sortByHeader = this.getColumnHeaders().find(h => h.id === sortBy);
    const sortedDesc = this.state.sortedDesc !== null ?
      this.state.sortedDesc :
      defaultSortDesc(sortByHeader);

    let nodes = getSortedNodes(this.props.nodes, sortByHeader, sortedDesc);
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
  sortedDesc: null,
  sortBy: null,
};
