import React from 'react';

export default class Plugins extends React.Component {
  renderPlugin(plugin) {
    return (
      <span className="plugins-plugin" key={plugin.id} title={plugin.description}>
        {plugin.label || plugin.id}
      </span>
    );
  }

  render() {
    const hasPlugins = this.props.plugins && this.props.plugins.length > 0;
    return (
      <div className="plugins">
        <span className="plugins-label">
          Plugins:
        </span>
        {hasPlugins && this.props.plugins.map((plugin, index) => this.renderPlugin(plugin, index))}
        {!hasPlugins && <span className="plugins-empty">n/a</span>}
      </div>
    );
  }
}
