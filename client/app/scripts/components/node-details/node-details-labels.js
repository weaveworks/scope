import React from 'react';

export default function NodeDetailsLabels({rows}) {
  return (
    <div className="node-details-labels">
      {rows.map(field => (<div className="node-details-labels-field" key={field.id}>
          <div className="node-details-labels-field-label truncate" title={field.label}>
            {field.label}
          </div>
          <div className="node-details-labels-field-value truncate" title={field.value}>
            {field.value}
          </div>
        </div>
      ))}
    </div>
  );
}
