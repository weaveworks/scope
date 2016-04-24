import _ from 'lodash';
import React from 'react';
import { Map as makeMap } from 'immutable';

import ShowMore from '../show-more';
import NodeDetailsTableRow from './node-details-table-row';
import { sortOrderChanged } from '../../actions/app-actions';


function isNumberField(field) {
  return field.dataType && field.dataType === 'number';
}


const COLUMN_WIDTHS = {
  port: '44px',
  count: '70px'
};

function getDefaultSortBy(columns, nodes) {
  // default sorter specified by columns
  const defaultSortColumn = _.find(columns, {defaultSort: true});
  if (defaultSortColumn) {
    return defaultSortColumn.id;
  }
  // otherwise choose first metric
  return _.get(nodes, [0, 'metrics', 0, 'id']);
}


function getValueForSortBy(sortBy) {
  // return the node's value based on the sortBy field
  return (node) => {
    if (sortBy !== null) {
      const field = _.union(node.metrics, node.metadata).find(f => f.id === sortBy);
      if (field) {
        if (isNumberField(field)) {
          return parseFloat(field.value);
        }
        return field.value;
      }
    }
    return -1e-10; // just under 0 to treat missing values differently from 0
  };
}


/*
function getMetaDataSorters(nodes) {
  // returns an array of sorters that will take a node
  return _.get(nodes, [0, 'metadata'], []).map((field, index) => node => {
    const nodeMetadataField = node.metadata[index];
    if (nodeMetadataField) {
      if (isNumberField(nodeMetadataField)) {
        return parseFloat(nodeMetadataField.value);
      }
      return nodeMetadataField.value;
    }
    return null;
  });
}
*/


function getSortedNodes(nodes, columns, sortBy, sortedDesc) {
  const sortedNodes = _.sortBy(
    nodes,
    getValueForSortBy(sortBy || getDefaultSortBy(columns, nodes)),
    'label'
    // getMetaDataSorters(nodes)
  );
  if (sortedDesc) {
    sortedNodes.reverse();
  }
  return sortedNodes;
}


export default class NodeDetailsTable extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.DEFAULT_LIMIT = 5;
    this.state = {
      limit: props.limit || this.DEFAULT_LIMIT,
      sortedDesc: true,
      sortBy: null
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
  }

  handleHeaderClick(ev, headerId) {
    ev.preventDefault();
    const sortedDesc = headerId === this.state.sortBy
      ? !this.state.sortedDesc : this.state.sortedDesc;
    const sortBy = headerId;
    this.setState({sortBy, sortedDesc});
    sortOrderChanged({sortBy, sortedDesc});
  }

  handleLimitClick() {
    const limit = this.state.limit ? 0 : this.DEFAULT_LIMIT;
    this.setState({limit});
  }

  renderHeaders() {
    if (this.props.nodes && this.props.nodes.length > 0) {
      const columns = this.props.columns || [];
      const headers = [{id: 'label', label: this.props.label}].concat(columns);
      const defaultSortBy = getDefaultSortBy(this.props);

      // Beauty hack: adjust first column width if there are only few columns;
      // this assumes the other columns are narrow metric columns of 20% table width
      if (headers.length === 2) {
        headers[0].width = '66%';
      } else if (headers.length === 3) {
        headers[0].width = '50%';
      } else if (headers.length >= 3 && headers.length < 5) {
        headers[0].width = '33%';
      }

      //
      // More beauty hacking, ports and counts can only get so big, free up WS for other longer
      // fields like IPs!
      //
      headers.forEach(h => {
        h.width = COLUMN_WIDTHS[h.id];
      });

      return (
        <tr>
          {headers.map(header => {
            const headerClasses = ['node-details-table-header', 'truncate'];
            const onHeaderClick = ev => {
              this.handleHeaderClick(ev, header.id);
            };

            // sort by first metric by default
            const isSorted = this.state.sortBy !== null
              ? header.id === this.state.sortBy : header.id === defaultSortBy;
            const isSortedDesc = isSorted && this.state.sortedDesc;
            const isSortedAsc = isSorted && !isSortedDesc;
            if (isSorted) {
              headerClasses.push('node-details-table-header-sorted');
            }

            // set header width in percent
            const style = {};
            if (header.width) {
              style.width = header.width;
            }

            return (
              <td className={headerClasses.join(' ')} style={style} onClick={onHeaderClick}
                key={header.id}>
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
    const headers = this.renderHeaders();
    const { nodeIdKey, columns, topologyId, onMouseOverRow } = this.props;
    let nodes = getSortedNodes(this.props.nodes, this.props.columns, this.state.sortBy,
                                    this.state.sortedDesc);
    const limited = nodes && this.state.limit > 0 && nodes.length > this.state.limit;
    const expanded = this.state.limit === 0;
    const notShown = nodes.length - this.state.limit;
    if (nodes && limited) {
      nodes = nodes.slice(0, this.state.limit);
    }

    const nodeOrderJS = (nodes || []).map((n, i) => [n.id, i]);
    const nodeOrder = makeMap(nodeOrderJS);
    const childrenWithProps = React.Children.map(this.props.children, (child) => (
      React.cloneElement(child, { nodeOrder })
    ));

    console.log(this.props.selectedRowId);

    return (
      <div className="node-details-table-wrapper-wrapper" style={this.props.style}>
        <div className="node-details-table-wrapper" onMouseOut={this.props.onMouseOut}>
          <table className="node-details-table">
            <thead>
              {headers}
            </thead>
            <tbody>
              {nodes && nodes.map(node => (
                <NodeDetailsTableRow
                  key={node.id}
                  selected={this.props.highlightedNodeIds.has(node.id)}
                  node={node}
                  nodeIdKey={nodeIdKey}
                  columns={columns}
                  onMouseOverRow={onMouseOverRow}
                  topologyId={topologyId} />
              ))}
            </tbody>
          </table>
          <ShowMore
            handleClick={this.handleLimitClick}
            collection={this.props.nodes}
            expanded={expanded}
            notShown={notShown} />
        </div>
        {childrenWithProps}
      </div>
    );
  }
}


NodeDetailsTable.defaultProps = {
  nodeIdKey: 'id' // key to identify a node in a row (used for topology links)
};
