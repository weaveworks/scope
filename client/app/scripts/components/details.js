import React from 'react';

import DetailsCard from './details-card';

export default class Details extends React.Component {

  // render all details as cards, later cards go on top
  render() {
    const details = this.props.details.toIndexedSeq();
    return (
      <div className="details">
        {details.map((obj, index) => {
          return (
            <DetailsCard key={obj.id} index={index} cardCount={details.size}
              nodes={this.props.nodes}
              nodeControlStatus={this.props.controlStatus[obj.id]} {...obj} />
          );
        })}
      </div>
    );
  }
}
