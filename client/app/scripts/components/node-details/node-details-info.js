import React from 'react';

export default class NodeDetailsInfo extends React.Component {
  render() {
    return (
      <div className="node-details-info">
        {this.props.rows && this.props.rows.map(field => {
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
      </div>
    );
  }
}
