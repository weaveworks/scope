import React from 'react';

import DetailsCard from './details-card';

export default function Details({controlStatus, details, nodes}) {
  // render all details as cards, later cards go on top
  return (
    <div className="details">
      {details.toIndexedSeq().map((obj, index) => <DetailsCard key={obj.id}
        index={index} cardCount={details.size} nodes={nodes}
        nodeControlStatus={controlStatus.get(obj.id)} {...obj} />
      )}
    </div>
  );
}
