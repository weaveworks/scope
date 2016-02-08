import React from 'react';

export default class NodeDetailsInfo extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      expanded: false
    };
    this.handleClickMore = this.handleClickMore.bind(this);
  }

  handleClickMore(ev) {
    ev.preventDefault();
    const expanded = !this.state.expanded;
    this.setState({expanded});
  }

  render() {
    const rows = (this.props.rows || []);
    const prime = rows.filter(row => row.prime);
    const overflow = rows.filter(row => !row.prime);
    const showOverflow = overflow.length > 0 && !this.state.expanded;
    const showLess = this.state.expanded;
    return (
      <div className="node-details-info">
        {prime && prime.map(field => {
          return (
            <div className="node-details-info-field" key={field.id}>
              <div className="node-details-info-field-label truncate" title={field.label}>
                {field.label}
              </div>
              <div className="node-details-info-field-value" title={field.value}>
                <div className="truncate">
                  {field.value}
                </div>
              </div>
            </div>
          );
        })}
        {this.state.expanded && overflow && overflow.map(field => {
          return (
            <div className="node-details-info-field" key={field.id}>
              <div className="node-details-info-field-label truncate" title={field.label}>
                {field.label}
              </div>
              <div className="node-details-info-field-value" title={field.value}>
                <div className="truncate">
                  {field.value}
                </div>
              </div>
            </div>
          );
        })}
        {showOverflow && <div className="node-details-info-overflow-expand" onClick={this.handleClickMore}>Show more</div>}
        {showLess && <div className="node-details-info-expand" onClick={this.handleClickMore}>Show less</div>}
      </div>
    );
  }
}
