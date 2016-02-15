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
    let rows = (this.props.rows || []);
    const prime = rows.filter(row => (this.props.show_all || row.prime));
    let expandText = 'Show less';
    let showExpand = this.state.expanded;
    if (!this.state.expanded && prime.length < rows.length) {
      expandText = 'Show more';
      showExpand = true;
      rows = prime;
    }
    return (
      <div className="node-details-info">
        {rows.map(field => {
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
        {showExpand && <div className="node-details-info-expand" onClick={this.handleClickMore}>{expandText}</div>}
      </div>
    );
  }
}
