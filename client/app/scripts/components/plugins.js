import React from 'react';

export default class Plugins extends React.Component {
  renderPlugin(plugin) {
    return (
      <div className="plugin" key={plugin.id} title={plugin.description}>
        {plugin.label || plugin.id}
      </div>
    );
  }

  render() {
    if (!this.props.plugins || this.props.plugins.length === 0) {
      return <div className="plugins">No plugins loaded</div>;
    }
    return (
      <div className="plugins">
        Plugins: {this.props.plugins.map(plugin => this.renderPlugin(plugin))}
      </div>
    );
  }
}
