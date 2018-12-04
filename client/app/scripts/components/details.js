import React from 'react';
import { connect } from 'react-redux';

import DetailsCard from './details-card';

class Details extends React.Component {
  render() {
    const { controlStatus, details } = this.props;
    // render all details as cards, later cards go on top
    return (
      <div className="details">
        {details.toIndexedSeq().map((obj, index) => (
          <DetailsCard
            key={obj.id}
            index={index}
            cardCount={details.size}
            nodeControlStatus={controlStatus.get(obj.id)}
            renderNodeDetailsExtras={this.props.renderNodeDetailsExtras}
            {...obj}
          />
        ))}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    controlStatus: state.get('controlStatus'),
    details: state.get('nodeDetails')
  };
}

export default connect(mapStateToProps)(Details);
