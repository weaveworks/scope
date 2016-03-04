import React from 'react';

import ShowMore from '../show-more';

export default class NodeDetailsInfo extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      expanded: false
    };
    this.handleClickMore = this.handleClickMore.bind(this);
  }

  handleClickMore() {
    const expanded = !this.state.expanded;
    this.setState({expanded});
  }

  render() {
    let rows = (this.props.rows || []);
    const prime = rows.filter(row => row.priority < 10);
    let notShown = 0;
    if (!this.state.expanded && prime.length < rows.length) {
      notShown = rows.length - prime.length;
      rows = prime;
    }
    return (
      <div className="node-details-info">
        {rows.map(field => (<div className="node-details-info-field" key={field.id}>
            <div className="node-details-info-field-label truncate" title={field.label}>
              {field.label}
            </div>
            <div className="node-details-info-field-value truncate" title={field.value}>
              {field.value}
            </div>
          </div>
        ))}
        <ShowMore handleClick={this.handleClickMore} collection={this.props.rows}
          expanded={this.state.expanded} notShown={notShown} />
      </div>
    );
  }
}
