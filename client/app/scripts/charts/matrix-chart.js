var d3 = require('d3');
var _ = require('lodash');
var EventEmitter = require('events').EventEmitter;


var MAX_NODES = 100;

function determineAllEdges(nodes) {
    var edges = [],
        edgeIds = {};

    _.each(nodes, function(node) {
        _.each(node.adjacency, function(adjacent) {
            var edge = [node.id, adjacent],
                edgeId = edge.join('-');

            if (!edgeIds[edgeId]) {
                edges.push({
                    id: edgeId,
                    value: 1,
                    source: nodes[edge[0]].index,
                    target: nodes[edge[1]].index
                });
                edgeIds[edgeId] = true;
            }
        });
    });

    return edges;
}

function matrixChart() {

    var preview = false,
        margin = {top: 0, right: 0, bottom: 0, left: 0},
        width = 0,
        height = 0;

    var dispatcher = new EventEmitter();

    var x = d3.scale.ordinal();

    function chart(selection) {
        selection.each(function(data) {

            if (_.size(data.nodes) > MAX_NODES) {
                console.log('too many nodes for matrix');
                return;
            }

            width = Math.min(width, height);
            height = Math.min(width, height);
            x.rangeBands([0, width]);

            // Select the svg element, if it exists.
            var svg = d3.select(this)
                .selectAll("svg")
                .data([data]);

            svg.enter()
                .append("svg")
                .append("rect")
                    .attr("class", "background");

            svg
                .attr('width', width)
                .attr('height', height)

            if (!preview) {
              svg.select('.background')
                  .attr("width", width)
                  .attr("height", height);
            }

          var matrix = [],
              nodes = _.values(data.nodes),
              n = nodes.length,
              tooDense = n > width;

          if (tooDense) {
              console.log('matrix too dense for width', nodes.length, width);
          }

          // Compute index per node.
          nodes.forEach(function(node, i) {
              node.index = i;
              matrix[i] = d3.range(n).map(function(j) { return {x: j, y: i, z: 0}; });
          });

          // Convert links to matrix; count character occurrences.
          var links = determineAllEdges(data.nodes);

          links.forEach(function(link) {
              matrix[link.source][link.target].z += link.value;
          });

          // Precompute the orders.
          var orders = {
              major: d3.range(n).sort(function(a, b) { return d3.ascending(nodes[a].label_major, nodes[b].label_major); }),
              minor: d3.range(n).sort(function(a, b) { return d3.ascending(nodes[a].label_minor, nodes[b].label_minor); })
          };

          // The default sort order.
          x.domain(orders.major);

          var c = d3.scale.category20();

          var row = svg.selectAll(".matrix-row")
              .data(matrix, function(d, i) {
                  return i;
              });

          var rowEnter = row.enter()
              .append("g")
                  .attr("class", "matrix-row");

          if (!tooDense) {
              rowEnter
                  .append('line')
                      .attr('class', 'matrix-grid');

              var textEnter = rowEnter.append("text")
                  .attr('class', 'hide')
                  .attr("text-anchor", "end")

              textEnter.append('tspan')
                  .attr("x", -4)
                  .attr("dy", "-24")
                  .attr('class', 'direction')
                  .text("Source");

              textEnter.append('tspan')
                  .attr("x", -4)
                  .attr("dy", "10")
                  .attr('class', 'major');

              textEnter.append('tspan')
                  .attr("x", -4)
                  .attr("dy", "10")
                  .attr('class', 'minor');

              rowEnter
                  .append('line')
                      .attr('class', 'matrix-label-line hide')
                      .attr('x1', -2);
            }

            row.exit()
                .remove();

            row.attr("transform", function(d, i) { return "translate(0," + x(i) + ")"; })
                .each(cells);

            row.select('line.matrix-grid')
                .attr('x1', 0)
                .attr('x2', width);

            row.select('.major')
                .text(function(d, i) { return nodes[i].label_major; });

            row.select('.minor')
                .text(function(d, i) { return nodes[i].label_minor; });

            row.select(".matrix-label-line")
                .attr("y1", x.rangeBand() / 2)
                .attr("y2", x.rangeBand() / 2);

            row.select("text")
                .attr("y", x.rangeBand() / 2);

            var column = svg.selectAll(".matrix-column")
                .data(matrix, function(d, i) {
                    return i;
                });

            var columnEnter = column.enter()
                .append("g")
                    .attr("class", "matrix-column");

            if (!tooDense) {
                columnEnter.append("line")
                    .attr('class', 'matrix-grid');

                textEnter = columnEnter.append("text")
                    .attr('class', 'hide');

                textEnter.append('tspan')
                    .attr("x", 4)
                    .attr("dy", "-30")
                    .attr('class', 'direction')
                    .text("Destination");

                textEnter.append('tspan')
                    .attr("x", 4)
                    .attr("dy", "10")
                    .attr('class', 'major');

                textEnter.append('tspan')
                    .attr("x", 4)
                    .attr("dy", "10")
                    .attr('class', 'minor');

                columnEnter
                    .append('line')
                        .attr('class', 'matrix-label-line hide')
                        .attr('y1', -2);

            }

            column.exit()
                .remove();

            column
                .attr("transform", function(d, i) { return "translate(" + x(i) + ")"; });

            column.select("line")
                .attr("y2", width);

            column.select("text")
                .attr("transform", "translate(" + (x.rangeBand() / 2) + ")");

            column.select(".matrix-label-line")
                .attr('y2', -40)
                .attr("x1", x.rangeBand() / 2)
                .attr("x2", x.rangeBand() / 2);

            column.select('.major')
                .text(function(d, i) { return nodes[i].label_major; });

            column.select('.minor')
                .text(function(d, i) { return nodes[i].label_minor; });

            function mouseover(p) {
                d3.selectAll("#matrix .matrix-row text").classed("hide", function(d, i) {
                    return i != p.y;
                });

                d3.selectAll("#matrix .matrix-row .matrix-label-line").classed("hide", function(d, i) {
                    return i != p.y;
                });

                d3.selectAll("#matrix .matrix-column text").classed("hide", function(d, i) {
                    return i != p.x;
                });

                d3.selectAll("#matrix .matrix-column .matrix-label-line").classed("hide", function(d, i) {
                    return i != p.x;
                });

                d3.selectAll('#matrix .matrix-row').each(function(d, i) {
                    if (i == p.y) {
                      var row = d3.select(this);

                      // determine text lengths
                      row.selectAll('tspan')
                          .each(function() {
                              p.textLength = Math.max(this.getComputedTextLength(), p.textLength || 0);
                          });

                      // set line length
                      row.select('.matrix-label-line')
                          .attr('x2', -p.textLength - 8);
                    }
                });

            }

            function click(p) {
                dispatcher.emit('node.click', nodes[p.y].id);
            }

            function cells(row) {
                var cell = d3.select(this).selectAll(".matrix-cell")
                    .data(row.filter(function(d){
                        return d.z;
                    }));

                var cellEnter = cell.enter().append("rect")
                    .attr("class", "matrix-cell");

                if (!tooDense) {
                    cellEnter.on('click', click)
                        .on("mouseover", mouseover);
                }

                cell.exit()
                    .remove();

                cell
                    .attr("x", function(d) { return x(d.x); })
                    .attr("width", x.rangeBand())
                    .attr("height", x.rangeBand())
                    .style("fill", function(d) {
                        return d.z ? c([nodes[d.x].label_major, nodes[d.y].label_major].sort().join('-')) : 'none';
                    });
            }

        });
    }

    chart.create = function(el, state) {
        d3.select(el)
            .datum(state)
            .call(chart);

        return chart;
    };

    chart.update = _.throttle(function(el, state) {
        d3.select(el)
            .datum(state)
            .call(chart);

        return chart;
    }, 500);

    chart.on = function(event, callback) {
        dispatcher.on(event, callback);
        return chart;
    };

    chart.preview = function(_) {
        if (!arguments.length) return preview;
        preview = _;
        return chart;
    };

    chart.width = function(_) {
        if (!arguments.length) return width;
        width = _;
        return chart;
    };

    chart.height = function(_) {
        if (!arguments.length) return height;
        height = _;
        return chart;
    };

    return chart;
}

module.exports = matrixChart;