import React from 'react';

import ShowMore from '../show-more';

export default class NodeDetailsLabels extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.DEFAULT_LIMIT = 5;
    this.state = {
      limit: this.DEFAULT_LIMIT,
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
  }

  handleLimitClick() {
    const limit = this.state.limit ? 0 : this.DEFAULT_LIMIT;
    this.setState({limit});
  }

  render() {
    let rows = this.props.rows;
    const limited = rows && this.state.limit > 0 && rows.length > this.state.limit;
    const expanded = this.state.limit === 0;
    const notShown = rows.length - this.DEFAULT_LIMIT;
    if (rows && limited) {
      rows = rows.slice(0, this.state.limit);
    }

    return (
      <div className="node-details-labels">
        {rows.map(field => (<div className="node-details-labels-field" key={field.id}>
            <div className="node-details-labels-field-label truncate" title={field.label}
              key={field.id}>
              {field.label}
            </div>
            <div className="node-details-labels-field-value truncate" title={field.value}>
              {field.value}
            </div>
          </div>
        ))}
        <ShowMore handleClick={this.handleLimitClick} collection={this.props.rows}
          expanded={expanded} notShown={notShown} />
      </div>
    );
  }
}
