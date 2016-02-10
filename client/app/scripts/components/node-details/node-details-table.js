import _ from 'lodash';
import React from 'react';

import NodeDetailsTableNodeLink from './node-details-table-node-link';
import NodeDetailsTableNodeMetric from './node-details-table-node-metric';

export default class NodeDetailsTable extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.DEFAULT_LIMIT = 5;
    this.state = {
      limit: this.DEFAULT_LIMIT,
      sortedDesc: true,
      sortBy: null
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
    this.getValueForSortBy = this.getValueForSortBy.bind(this);
  }

  handleHeaderClick(ev, headerId) {
    ev.preventDefault();
    const sortedDesc = headerId === this.state.sortBy ? !this.state.sortedDesc : this.state.sortedDesc;
    const sortBy = headerId;
    this.setState({sortBy, sortedDesc});
  }

  handleLimitClick(ev) {
    ev.preventDefault();
    const limit = this.state.limit ? 0 : this.DEFAULT_LIMIT;
    this.setState({limit});
  }

  getDefaultSortBy() {
    // first metric
    return _.get(this.props.nodes, [0, 'metrics', 0, 'id']);
  }

  getMetaDataSorters() {
    // returns an array of sorters that will take a node
    return _.get(this.props.nodes, [0, 'metadata'], []).map((field, index) => {
      return node => node.metadata[index] ? node.metadata[index].value : null;
    });
  }

  getValueForSortBy(node) {
    // return the node's value based on the sortBy field
    const sortBy = this.state.sortBy || this.getDefaultSortBy();
    if (sortBy !== null) {
      const field = _.union(node.metrics, node.metadata).find(f => f.id === sortBy);
      if (field) {
        return field.value;
      }
    }
    return -1e-10; // just under 0 to treat missing values differently from 0
  }

  getValuesForNode(node) {
    const values = {};
    ['metrics', 'metadata'].forEach(collection => {
      if (node[collection]) {
        node[collection].forEach(field => {
          field.valueType = collection;
          values[field.id] = field;
        });
      }
    });
    return values;
  }

  renderHeaders() {
    if (this.props.nodes && this.props.nodes.length > 0) {
      const headers = [{id: 'label', label: this.props.label}].concat(this.props.columns);
      const defaultSortBy = this.getDefaultSortBy();

      return (
        <tr>
          {headers.map(header => {
            const headerClasses = ['node-details-table-header', 'truncate'];
            const onHeaderClick = ev => {
              this.handleHeaderClick(ev, header.id);
            };
            // sort by first metric by default
            const isSorted = this.state.sortBy !== null ? header.id === this.state.sortBy : header.id === defaultSortBy;
            const isSortedDesc = isSorted && this.state.sortedDesc;
            const isSortedAsc = isSorted && !isSortedDesc;
            if (isSorted) {
              headerClasses.push('node-details-table-header-sorted');
            }
            return (
              <td className={headerClasses.join(' ')} onClick={onHeaderClick} key={header.id}>
                {isSortedAsc && <span className="node-details-table-header-sorter fa fa-caret-up" />}
                {isSortedDesc && <span className="node-details-table-header-sorter fa fa-caret-down" />}
                {header.label}
              </td>
            );
          })}
        </tr>
      );
    }
    return '';
  }

  renderValues(node) {
    const fields = this.getValuesForNode(node);
    return this.props.columns.map(({id}) => {
      const field = fields[id];
      if (field) {
        if (field.valueType === 'metadata') {
          return (
            <td className="node-details-table-node-value" key={field.id}>
              {field.value}
            </td>
          );
        }
        return <NodeDetailsTableNodeMetric key={field.id} {...field} />;
      }
    });
  }

  render() {
    const headers = this.renderHeaders();
    let nodes = _.sortByAll(this.props.nodes, this.getValueForSortBy, 'label', this.getMetaDataSorters());
    const limited = nodes && this.state.limit > 0 && nodes.length > this.state.limit;
    const showLimitAction = nodes && (limited || (this.state.limit === 0 && nodes.length > this.DEFAULT_LIMIT));
    const limitActionText = limited ? 'Show more' : 'Show less';
    if (this.state.sortedDesc) {
      nodes.reverse();
    }
    if (nodes && limited) {
      nodes = nodes.slice(0, this.state.limit);
    }

    return (
      <div className="node-details-table-wrapper">
        <table className="node-details-table">
          <thead>
            {headers}
          </thead>
          <tbody>
          {nodes && nodes.map(node => {
            const values = this.renderValues(node);
            return (
              <tr className="node-details-table-node" key={node.id}>
                <td className="node-details-table-node-label truncate">
                  <NodeDetailsTableNodeLink topologyId={this.props.topologyId} {...node} />
                </td>
                {values}
              </tr>
            );
          })}
          </tbody>
        </table>
        {showLimitAction && <div className="node-details-table-more" onClick={this.handleLimitClick}>{limitActionText}</div>}
      </div>
    );
  }
}
