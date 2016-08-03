import React from 'react';

const GENERAL_SHORTCUTS = [
  {key: 'esc', label: 'Close active panel'},
  {key: '/', label: 'Activate search field'},
  {key: '?', label: 'Toggle shortcut menu'},
  {key: 't', label: 'Toggle Table mode'},
  {key: 'g', label: 'Toggle Graph mode'},
];

const CANVAS_METRIC_SHORTCUTS = [
  {key: '<', label: 'Select and pin previous metric'},
  {key: '>', label: 'Select and pin next metric'},
  {key: 'q', label: 'Unpin current metric'},
];

function renderShortcuts(cuts) {
  return (
    <div>
      {cuts.map(({key, label}) => (
        <div key={key} className="help-panel-shortcut">
          <div className="key"><kbd>{key}</kbd></div>
          <div className="label">{label}</div>
        </div>
      ))}
    </div>
  );
}

export default class HelpPanel extends React.Component {
  render() {
    return (
      <div className="help-panel">
        <div className="help-panel-header">
          <h2>Keyboard Shortcuts</h2>
        </div>
        <div className="help-panel-main">
          <h3>General</h3>
          {renderShortcuts(GENERAL_SHORTCUTS)}
          <h3>Canvas Metrics</h3>
          {renderShortcuts(CANVAS_METRIC_SHORTCUTS)}
        </div>
      </div>
    );
  }
}
