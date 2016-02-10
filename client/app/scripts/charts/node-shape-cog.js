import React from 'react';

const COG_PATH = 'M61.321,26.585l-5.292-1.769c-0.645-2.729-1.718-5.295-3.15-7.615l2.486-4.989c-1.706-2.251-3.717-4.26-5.968-5.968l-4.989,2.488c-2.321-1.432-4.885-2.508-7.615-3.152l-1.767-5.289C33.643,0.102,32.235,0,30.806,0c-1.435,0-2.84,0.102-4.221,0.291L24.816,5.58c-2.731,0.645-5.297,1.721-7.618,3.152l-4.986-2.488c-2.251,1.708-4.263,3.717-5.968,5.968l2.485,4.989c-1.432,2.32-2.505,4.887-3.152,7.615l-5.286,1.769C0.099,27.964,0,29.371,0,30.806c0,1.433,0.099,2.84,0.291,4.219l5.286,1.771c0.647,2.729,1.721,5.295,3.152,7.616l-2.485,4.986c1.705,2.251,3.717,4.263,5.968,5.97l4.986-2.488c2.321,1.435,4.887,2.508,7.618,3.152l1.766,5.289c1.384,0.189,2.789,0.289,4.222,0.289c1.432,0,2.839-0.1,4.221-0.289l1.769-5.289c2.73-0.645,5.294-1.718,7.615-3.152l4.986,2.488c2.254-1.707,4.265-3.719,5.971-5.97l-2.488-4.986c1.435-2.321,2.508-4.888,3.152-7.616l5.289-1.771c0.191-1.379,0.291-2.786,0.291-4.219C61.609,29.371,61.51,27.964,61.321,26.585z';

export default function NodeShapeCog({highlighted, size, color}) {
  const width = -61.609;
  const height = -61.609;
  const cx = Math.abs(width) / 2;
  const cy = Math.abs(height) / 2;
  const pathSize = Math.abs((width + height) / 2);
  const baseScale = (size * 2.1) / pathSize;

  const pathProps = (v) => {
    return {
      d: COG_PATH,
      transform: `scale(-${v * baseScale}) translate(-${cx},-${cy})`,
      style: {strokeWidth: 3 / baseScale}
    };
  };

  return (
    <g className="shape">
      {highlighted &&
        <circle r={size * 0.7} className="highlighted"></circle>}
      <path className="border" stroke={color} {...pathProps(0.5)} />
      <path className="shadow" {...pathProps(0.45)} />
      <circle className="node" r={Math.max(2, (size * 0.125))} />
    </g>
  );
}
