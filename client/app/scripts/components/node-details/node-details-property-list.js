import React from 'react';
import { Map as makeMap } from 'immutable';
import sortBy from 'lodash/sortBy';

import { NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT } from '../../constants/limits';
import NodeDetailsControlButton from './node-details-control-button';
import MatchedText from '../matched-text';
import ShowMore from '../show-more';

const Controls = controls => (
  <div className="node-details-property-list-controls">
    {sortBy(controls, 'rank').map(control => (<NodeDetailsControlButton
      nodeId={control.nodeId}
      control={control}
      key={control.id} />))}
  </div>
);

export default class NodeDetailsPropertyList extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.state = {
      limit: NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT,
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
  }

  handleLimitClick() {
    const limit = this.state.limit ? 0 : NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT;
    this.setState({limit});
  }

  render() {
    const { controls, matches = makeMap() } = this.props;
    let { rows } = this.props;
    let notShown = 0;
    const limited = rows && this.state.limit > 0 && rows.length > this.state.limit;
    const expanded = this.state.limit === 0;
    if (rows && limited) {
      const hasNotShownMatch = rows.filter((row, index) => index >= this.state.limit
        && matches.has(row.id)).length > 0;
      if (!hasNotShownMatch) {
        notShown = rows.length - NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT;
        rows = rows.slice(0, this.state.limit);
      }
    }

    return (
      <div className="node-details-property-list">
        {controls && Controls(controls)}
        {rows.map(field => (
          <div className="node-details-property-list-field" key={field.id}>
            <div
              className="node-details-property-list-field-label truncate"
              title={field.entries.label}
              key={field.id}>
              {field.entries.label}
            </div>
            <div
              className="node-details-property-list-field-value truncate"
              title={field.entries.value}>
              <MatchedText text={field.entries.value} match={matches.get(field.id)} />
            </div>
          </div>
        ))}
        <ShowMore
          handleClick={this.handleLimitClick}
          collection={this.props.rows}
          expanded={expanded}
          notShown={notShown} />
      </div>
    );
  }
}
