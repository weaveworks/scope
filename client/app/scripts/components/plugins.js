import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

class Plugins extends React.Component {
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
    const hasPlugins = this.props.plugins && this.props.plugins.size > 0;
    return (
      <div className="plugins">
        <span className="plugins-label">
          Plugins:
        </span>
        {hasPlugins && this.props.plugins.toIndexedSeq()
          .map(plugin => this.renderPlugin(plugin.toJS()))}
        {!hasPlugins && <span className="plugins-empty">n/a</span>}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    plugins: state.get('plugins')
  };
}

export default connect(
  mapStateToProps
)(Plugins);
