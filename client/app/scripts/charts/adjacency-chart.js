var d3 = require('d3');
var _ = require('lodash');
var EventEmitter = require('events').EventEmitter;


var MAX_NODES = 100;

function filterNodes(nodes) {
    return _.filter(nodes, function(n) {
        return !n.pseudo;
    });
} 

function adjacencyChart() {

    var preview = false,
        margin = {top: 0, right: 0, bottom: 0, left: 0},
        width = 0,
        height = 0;

    var dispatcher = new EventEmitter();

    var x = d3.scale.ordinal(),
        c = d3.scale.category20();

    function chart(selection) {

        selection.each(function(data) {

            if (_.size(data.nodes) > MAX_NODES) {
                console.log('too many nodes for matrix');
                return;
            }

            width = Math.min(width, height);
            height = Math.min(width, height);

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
              nodeMap = data.nodes,
              nodes = _.sortBy(filterNodes(data.nodes), 'label_major'),
              n = nodes.length,
              tooDense = n > width;

          if (tooDense) {
              console.log('matrix too dense for width', nodes.length, width);
          }

          x.rangeBands([0, width])
              .domain(d3.range(n));


          // Compute index per node.
          nodes.forEach(function(node, i) {
              node.index = i;
              node.count = 0;
              matrix[i] = d3.range(n).map(function(j) { return {x: j, y: i, z: 0}; });
          });

          var row = svg.selectAll(".matrix-row")
              .data(nodes, function(d, i) {
                  return d.id;
              });

          var rowEnter = row.enter()
              .append("g")
                  .attr("class", "matrix-row");

          if (!tooDense) {
              rowEnter.append('line');
              rowEnter.append("text")
                  .classed('hide', function(d, i) {
                      return i != 0;
                  })
                  .attr("x", -6)
                  .attr("dy", ".32em")
                  .attr("text-anchor", "end");
            }

            row.exit()
                .remove();

            row
                .attr("transform", function(d, i) {
                    return "translate(0," + x(i) + ")";
                })
                .each(cells);

            row.select('line').attr('x2', width);

            row.select("text")
                .attr("y", x.rangeBand() / 2)
                .text(function(d, i) { return d.label_major; });

            var column = svg.selectAll(".matrix-column")
                .data(nodes, function(d, i) {
                    return d.id;
                });

            var columnEnter = column.enter()
                .append("g")
                    .attr("class", "matrix-column");

            if (!tooDense) {
                columnEnter.append("line");

                columnEnter.append("text")
                    .classed('hide', function(d, i) {
                        return i != 0;
                    })
                    .attr("x", 6)
                    .attr("dy", ".32em")
                    .attr("text-anchor", "start")
            }

            column.exit().remove();

            column
                .attr("transform", function(d, i) { return "translate(" + x(i) + ")rotate(-90)"; });

            column.select("line")
                .attr("x1", -width);

            function cells(row) {
                var cell = d3.select(this).selectAll(".matrix-cell")
                    .data(row.adjacency || []);

                var cellEnter = cell.enter().append("rect")
                    .attr("class", "matrix-cell");

                cell.exit()
                    .remove();

                cell
                    .attr("x", function(d, i) { return x(i); })
                    .attr("width", x.rangeBand())
                    .attr("height", x.rangeBand())
                    .style("fill", function(d) {
                        return c(nodeMap[d].label_major);
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

module.exports = adjacencyChart;