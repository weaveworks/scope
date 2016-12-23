import React from 'react';
import { Map as makeMap } from 'immutable';
import { sortBy } from 'lodash';

import MatchedText from '../matched-text';
import ShowMore from '../show-more';


function columnStyle(column) {
  return {
    textAlign: column.dataType === 'number' ? 'right' : 'left',
    paddingRight: '10px',
    maxWidth: '140px'
  };
}

function sortedRows(rows, sortedByColumn, sortedDesc) {
  const orderedRows = sortBy(rows, row => row.id);
  const sorted = sortBy(orderedRows, (row) => {
    let value = row.entries[sortedByColumn.id];
    if (sortedByColumn.dataType === 'number') {
      value = parseFloat(value);
    }
    return value;
  });
  if (!sortedDesc) {
    sorted.reverse();
  }
  return sorted;
}

export default class NodeDetailsGenericTable extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.DEFAULT_LIMIT = 5;
    this.state = {
      limit: this.DEFAULT_LIMIT,
      sortedByColumn: props.columns[0],
      sortedDesc: true
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
  }

  handleHeaderClick(ev, column) {
    ev.preventDefault();
    this.setState({
      sortedByColumn: column,
      sortedDesc: this.state.sortedByColumn.id === column.id
        ? !this.state.sortedDesc : true
    });
  }

  handleLimitClick() {
    const limit = this.state.limit ? 0 : this.DEFAULT_LIMIT;
    this.setState({limit});
  }

  render() {
    const { sortedByColumn, sortedDesc } = this.state;
    const { columns, matches = makeMap() } = this.props;
    let rows = this.props.rows;
    let notShown = 0;
    const limited = rows && this.state.limit > 0 && rows.length > this.state.limit;
    const expanded = this.state.limit === 0;
    if (rows && limited) {
      const hasNotShownMatch = rows.filter((row, index) => index >= this.state.limit
        && matches.has(row.id)).length > 0;
      if (!hasNotShownMatch) {
        notShown = rows.length - this.DEFAULT_LIMIT;
        rows = rows.slice(0, this.state.limit);
      }
    }

    return (
      <div className="node-details-generic-table">
        <table>
          <thead>
            <tr>
              {columns.map((column) => {
                const onHeaderClick = (ev) => {
                  this.handleHeaderClick(ev, column);
                };
                const isSorted = column.id === this.state.sortedByColumn.id;
                const isSortedDesc = isSorted && this.state.sortedDesc;
                const isSortedAsc = isSorted && !isSortedDesc;
                const style = Object.assign(columnStyle(column), {
                  cursor: 'pointer',
                  fontSize: '11px'
                });
                return (
                  <th
                    className="node-details-generic-table-header"
                    key={column.id} style={style} onClick={onHeaderClick}>
                    {isSortedAsc
                      && <span className="node-details-table-header-sorter fa fa-caret-up" />}
                    {isSortedDesc
                      && <span className="node-details-table-header-sorter fa fa-caret-down" />}
                    {column.label}
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {sortedRows(rows, sortedByColumn, sortedDesc).map(row => (
              <tr className="node-details-generic-table-row" key={row.id}>
                {columns.map((column) => {
                  const value = row.entries[column.id];
                  const match = matches.get(column.id);
                  return (
                    <td
                      className="node-details-generic-table-field-value truncate"
                      title={value} key={column.id} style={columnStyle(column)}>
                      <MatchedText text={value} match={match} />
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
        <ShowMore
          handleClick={this.handleLimitClick} collection={this.props.rows}
          expanded={expanded} notShown={notShown}
        />
      </div>
    );
  }
}
