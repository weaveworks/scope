import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import Tooltip from './tooltip';


const Plugin = ({
  id, label, description, status
}) => {
  const error = status !== 'ok';
  const className = classNames({ error });
  const tip = (<span>Description: {description}<br />Status: {status}</span>);

  // Inner span to hold styling so we don't effect the "before:content"
  return (
    <span className="plugins-plugin" key={id}>
      <Tooltip tip={tip}>
        <span className={className}>
          {error && <i className="plugins-plugin-icon fa fa-exclamation-circle" />}
          {label || id}
        </span>
      </Tooltip>
    </span>
  );
};

class Plugins extends React.Component {
  render() {
    const hasPlugins = this.props.plugins && this.props.plugins.size > 0;
    return (
      <div className="plugins">
        <span className="plugins-label">
          Plugins:
        </span>
        {hasPlugins && this.props.plugins.toIndexedSeq().map(plugin => Plugin(plugin.toJS()))}
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

export default connect(mapStateToProps)(Plugins);
