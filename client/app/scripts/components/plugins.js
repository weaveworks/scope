import React from 'react';
import classNames from 'classnames';

export default class Plugins extends React.Component {
  renderPlugin({id, label, description, status}) {
    const error = status !== 'ok';
    const className = classNames({ error });
    const title = `Status: ${status} | Plugin description: ${description}`;

    // Inner span to hold styling so we don't effect the "before:content"
    return (
      <span className="plugins-plugin" key={id}>
        <span className={className} title={title}>
          {error && <span className="plugins-plugin-icon fa fa-exclamation-circle" />}
          {label || id}
        </span>
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
