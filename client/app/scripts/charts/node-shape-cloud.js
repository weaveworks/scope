import React from 'react';
import d3 from 'd3';

const CLOUD_PATH = 'M 1920,384 Q 1920,225 1807.5,112.5 1695,0 1536,0 H 448 '
  + 'Q 263,0 131.5,131.5 0,263 0,448 0,580 71,689.5 142,799 258,853 '
  + 'q -2,28 -2,43 0,212 150,362 150,150 362,150 158,0 286.5,-88 128.5,-88 '
  + '187.5,-230 70,62 166,62 106,0 181,-75 75,-75 75,-181 0,-75 -41,-138 '
  + '129,-30 213,-134.5 84,-104.5 84,-239.5 z';

function toPoint(stringPair) {
  return stringPair.split(',').map(p => parseFloat(p, 10));
}

function getExtents(svgPath) {
  const points = svgPath.split(' ').filter(s => s.length > 1).map(toPoint);
  return [d3.extent(points, p => p[0]), d3.extent(points, p => p[1])];
}

export default function NodeShapeCloud({highlighted, size, color, darkColor}) {
  const [[minx, maxx], [miny, maxy]] = getExtents(CLOUD_PATH);
  const width = (maxx - minx);
  const height = (maxy - miny);
  const cx = width / 2;
  const cy = height / 2;
  const pathSize = (width + height) / 2;
  const baseScale = (size * 2) / pathSize;
  const strokeWidth = 5 / baseScale;
  const shadowColor = '#FFF';

  const pathProps = v => ({
    d: CLOUD_PATH,
    transform: `scale(-${v * baseScale}) translate(-${cx},-${cy})`,
    strokeWidth
  });

  return (
    <g className="shape shape-cloud">
      {highlighted &&
          <path className="highlighted" {...pathProps(0.7)} />}
      <path className="outline" stroke={darkColor} fill="none" {...pathProps(0.53)} />
      <path className="border" stroke={color} {...pathProps(0.5)} fill={shadowColor} />
      <circle className="node" r={Math.max(1.33333, (size * 0.08333))} fill={darkColor} />
    </g>
  );
}
