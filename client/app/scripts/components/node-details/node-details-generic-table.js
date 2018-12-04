import React from 'react';
import sortBy from 'lodash/sortBy';
import { Map as makeMap } from 'immutable';

import { NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT } from '../../constants/limits';

import {
  isNumeric,
  getTableColumnsStyles,
  genericTableEntryKey
} from '../../utils/node-details-utils';
import NodeDetailsTableHeaders from './node-details-table-headers';
import MatchedText from '../matched-text';
import ShowMore from '../show-more';


function sortedRows(rows, columns, sortedBy, sortedDesc) {
  const column = columns.find(c => c.id === sortedBy);
  const sorted = sortBy(rows, (row) => {
    let value = row.entries[sortedBy];
    if (isNumeric(column)) {
      value = parseFloat(value);
    }
    return value;
  });
  if (sortedDesc) {
    sorted.reverse();
  }
  return sorted;
}

export default class NodeDetailsGenericTable extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.state = {
      limit: NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT,
      sortedBy: props.columns && props.columns[0].id,
      sortedDesc: true
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
    this.updateSorted = this.updateSorted.bind(this);
  }

  updateSorted(sortedBy, sortedDesc) {
    this.setState({ sortedBy, sortedDesc });
  }

  handleLimitClick() {
    this.setState({
      limit: this.state.limit ? 0 : NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT
    });
  }

  render() {
    const { sortedBy, sortedDesc } = this.state;
    const { columns, matches = makeMap() } = this.props;
    const expanded = this.state.limit === 0;

    let rows = this.props.rows || [];
    let notShown = 0;

    // If there are rows that would be hidden behind 'show more', keep them
    // expanded if any of them match the search query; otherwise hide them.
    if (this.state.limit > 0 && rows.length > this.state.limit) {
      const hasHiddenMatch = rows.slice(this.state.limit).some(row =>
        columns.some(column => matches.has(genericTableEntryKey(row, column))));
      if (!hasHiddenMatch) {
        notShown = rows.length - NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT;
        rows = rows.slice(0, this.state.limit);
      }
    }

    const styles = getTableColumnsStyles(columns);
    return (
      <div className="node-details-generic-table">
        <table>
          <thead>
            <NodeDetailsTableHeaders
              headers={columns}
              sortedBy={sortedBy}
              sortedDesc={sortedDesc}
              onClick={this.updateSorted}
            />
          </thead>
          <tbody>
            {sortedRows(rows, columns, sortedBy, sortedDesc).map(row => (
              <tr className="node-details-generic-table-row" key={row.id}>
                {columns.map((column, index) => {
                  const match = matches.get(genericTableEntryKey(row, column));
                  const value = row.entries[column.id];
                  return (
                    <td
                      className="node-details-generic-table-value truncate"
                      title={value}
                      key={column.id}
                      style={styles[index]}>
                      {column.dataType === 'link' ?
                        <a
                          rel="noopener noreferrer"
                          target="_blank"
                          className="node-details-table-node-link"
                          href={value}>
                          {value}
                        </a> :
                        <MatchedText text={value} match={match} />
                      }
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
        <ShowMore
          handleClick={this.handleLimitClick}
          collection={this.props.rows}
          expanded={expanded}
          notShown={notShown}
        />
      </div>
    );
  }
}
