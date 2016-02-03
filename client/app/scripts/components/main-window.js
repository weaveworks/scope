import React from 'react';
import { DETAILS_PANEL_WIDTH, DETAILS_PANEL_MARGINS,
  DETAILS_PANEL_OFFSET } from '../constants/styles';

export default function MainWindow({children, cardsCount}) {
  const style = {
    right: DETAILS_PANEL_MARGINS.right + DETAILS_PANEL_WIDTH + 10 +
      (cardsCount * DETAILS_PANEL_OFFSET)
  };

  const childrenWithProps = React.Children.map(children, (child) => {
    return React.cloneElement(child, {containerMargin: style.right});
  });

  return (
    <div className="main-window" style={style}>
      {childrenWithProps}
    </div>
  );
}
