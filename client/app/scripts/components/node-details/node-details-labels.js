import React from 'react';

export default class NodeDetailsLabels extends React.Component {
  render() {
    return (
      <div className="node-details-labels">
        {this.props.rows.map(field => {
          return (
            <div className="node-details-labels-field" key={field.id}>
              <div className="node-details-labels-field-label truncate" title={field.label}>
                {field.label}
              </div>
              <div className="node-details-labels-field-value" title={field.value}>
                <div className="truncate">
                  {field.value}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    );
  }
}
