import React from 'react';
import { connect } from 'react-redux';


class TimeTravelButton extends React.Component {
  render() {
    const { className, onClick, isTimeTravelling, hasTimeTravel } = this.props;

    if (!hasTimeTravel) return null;

    return (
      <span className={className} onClick={onClick} title="Travel back in time">
        {isTimeTravelling && <span className="fa fa-clock-o" />}
        <span className="label">Time Travel</span>
      </span>
    );
  }
}

function mapStateToProps({ root }, { params }) {
  const cloudInstance = root.instances[params.orgId] || {};
  const featureFlags = cloudInstance.featureFlags || [];
  return {
    hasTimeTravel: featureFlags.includes('time-travel'),
  };
}

export default connect(mapStateToProps)(TimeTravelButton);
