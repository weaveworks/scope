import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';
import ReactTooltip from 'react-tooltip';


const Plugin = ({id, label, description, status}) => {
  const error = status !== 'ok';
  const className = classNames({ error });
  const title = `Plugin description: ${description}<br />Status: ${status}`;

  // Inner span to hold styling so we don't effect the "before:content"
  return (
    <span className="plugins-plugin" key={id}>
      <span className={className} data-tip={title} data-multiline>
        {error && <span className="plugins-plugin-icon fa fa-exclamation-circle" />}
        {label || id}
      </span>
      <ReactTooltip class="tooltip" effect="solid" offset={{right: 7}} />
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

export default connect(
  mapStateToProps
)(Plugins);
