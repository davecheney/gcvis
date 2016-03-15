package main

const (
	GCVIS_TMPL = `
<html>
<head>
<title>gcvis - {{ .Title }}</title>
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.selection.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.stack.min.js"></script>

<script type="text/javascript">

(function() {
	var datagraph_data = [
		{ label: "gc.heapinuse", data: {{ .HeapUse }} },
		{ label: "scvg.inuse", data: {{ .ScvgInuse }} },
		{ label: "scvg.idle", data: {{ .ScvgIdle }} },
		{ label: "scvg.sys", data: {{ .ScvgSys }} },
		{ label: "scvg.released", data: {{ .ScvgReleased }} },
		{ label: "scvg.consumed", data: {{ .ScvgConsumed }} }
	];

	var datagraph_options = {
		legend: {
			position: "nw",
			noColumns: 2,
			backgroundOpacity: 0.2
		},
		yaxis: {
			tickFormatter: function(val) { return val + "MB"; }
		},
		xaxis: {
			tickFormatter: function(val) { return val + "s"; }
		},
		selection: {
			mode: "x"
		},
	};

	var clockgraph_data = [
		{ label: "STW sweep clock", data: {{ .STWSclock }} },
		{ label: "con mas clock", data: {{ .MASclock }} },
		{ label: "STW mark clock", data: {{ .STWMclock }} },
	];
	var cpugraph_data = [
		{ label: "STW sweep cpu", data: {{ .STWScpu }} },
		{ label: "con mas assist cpu", data: {{ .MASAssistcpu }} },
		{ label: "con mas bg cpu", data: {{ .MASBGcpu }} },
		{ label: "con mas idle cpu", data: {{ .MASIdlecpu }} },
		{ label: "STW mark cpu", data: {{ .STWMcpu }} },
	];

	var timingsgraph_options = {
		legend: {
			position: "nw",
			noColumns: 2,
			backgroundOpacity: 0.2
		},
		yaxis: {
			tickFormatter: function(val) { return val + "ms"; }
		},
		xaxis: {
			tickFormatter: function(val) { return val + "s"; }
		},
		selection: {
			mode: "x"
		},
		series: {
			stack: 0,
			lines: {
				show: true,
				fill:true,
				lineWidth: 0,
			},
		},
	};

	$(document).ready(function() {
		var datagraph = $.plot("#datagraph", datagraph_data, datagraph_options);
		var clockgraph = $.plot("#clockgraph", clockgraph_data, timingsgraph_options);
		var cpugraph = $.plot("#cpugraph", cpugraph_data, timingsgraph_options);

		var overview = $.plot("#overview", {}, {
			legend: { show: false},
			series: {
				lines: {
					show: true,
					lineWidth: 1
				},
				shadowSize: 0
			},
			xaxis: {
				ticks: [],
				min: 0,
				autoscaleMargin: 0.1
			},
			yaxis: {
				ticks: [],
				min: 0,
				autoscaleMargin: 0.1
			},
			selection: {
				mode: "x"
			}
		});

		// now connect the four
		$("#datagraph").bind("plotselected", function (event, ranges) {

			// do the zooming
			$.each(datagraph.getXAxes(), function(_, axis) {
				var opts = axis.options;
				opts.min = ranges.xaxis.from;
				opts.max = ranges.xaxis.to;
			});
			datagraph.setupGrid();
			datagraph.draw();
			datagraph.clearSelection();

			// don't fire event on the overview to prevent eternal loop
			overview.setSelection(ranges, true);
			clockgraph.setSelection(ranges, true);
			cpugraph.setSelection(ranges, true);
		});

		$("#clockgraph").bind("plotselected", function (event, ranges) {

			// do the zooming
			$.each(clockgraph.getXAxes(), function(_, axis) {
				var opts = axis.options;
				opts.min = ranges.xaxis.from;
				opts.max = ranges.xaxis.to;
			});
			clockgraph.setupGrid();
			clockgraph.draw();
			clockgraph.clearSelection();

			// don't fire event on the overview to prevent eternal loop

			overview.setSelection(ranges, true);
			datagraph.setSelection(ranges, true);
			cpugraph.setSelection(ranges, true);
		});

		$("#cpugraph").bind("plotselected", function (event, ranges) {

			// do the zooming
			$.each(cpugraph.getXAxes(), function(_, axis) {
				var opts = axis.options;
				opts.min = ranges.xaxis.from;
				opts.max = ranges.xaxis.to;
			});
			cpugraph.setupGrid();
			cpugraph.draw();
			cpugraph.clearSelection();

			// don't fire event on the overview to prevent eternal loop

			overview.setSelection(ranges, true);
			datagraph.setSelection(ranges, true);
			clockraph.setSelection(ranges, true);
		});

		$("#overview").bind("plotselected", function (event, ranges) {
			datagraph.setSelection(ranges);
			clockgraph.setSelection(ranges);
			cpugraph.setSelection(ranges);
		});

		// refresh data every second
		pullAndRedraw();

		function pullAndRedraw() {
			$.get(window.location.href + 'graph.json', function(graphData) {
				var datagraph_data = [
					{ label: "gc.heapinuse", data: graphData.HeapUse },
					{ label: "scvg.inuse", data: graphData.ScvgInuse },
					{ label: "scvg.idle", data: graphData.ScvgIdle },
					{ label: "scvg.sys", data: graphData.ScvgSys },
					{ label: "scvg.released", data: graphData.ScvgReleased },
					{ label: "scvg.consumed", data: graphData.ScvgConsumed }
				];
				var clockgraph_data = [
					{ label: "STW sweep clock",    data: graphData.STWSclock },
					{ label: "con mas clock",      data: graphData.MASclock },
					{ label: "STW mark clock",     data: graphData.STWMclock },
				];
				var cpugraph_data = [
					{ label: "STW sweep cpu",      data: graphData.STWScpu },
					{ label: "con mas assist cpu", data: graphData.MASAssistcpu },
					{ label: "con mas bg cpu",     data: graphData.MASBGcpu },
					{ label: "con mas idle cpu",   data: graphData.MASIdlecpu },
					{ label: "STW mark cpu",       data: graphData.STWMcpu },
				];

				datagraph.setData(datagraph_data);
				datagraph.setupGrid();
				datagraph.draw();

				clockgraph.setData(clockgraph_data);
				clockgraph.setupGrid();
				clockgraph.draw();

				cpugraph.setData(cpugraph_data);
				cpugraph.setupGrid();
				cpugraph.draw();

				overview.setData(datagraph_data);
				overview.setupGrid();
				overview.draw();

				setTimeout(pullAndRedraw, 1000);
			})
		}
	});
})();
</script>
<style>
#content {
	margin: 0 auto;
	padding: 10px;
}

#export {
	float: right;
}
dt { float: left; font-weight:bold; width: 160px; }
dd { margin-left: 160px; }

.graph-container {
	box-sizing: border-box;
	width: 1200px;
	height: 340px;
	padding: 20px 15px 15px 15px;
	margin: 15px auto 30px auto;
	border: 1px solid #ddd;
	background: #fff;
	background: linear-gradient(#f6f6f6 0, #fff 50px);
	background: -o-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -ms-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -moz-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -webkit-linear-gradient(#f6f6f6 0, #fff 50px);
	box-shadow: 0 3px 10px rgba(0,0,0,0.15);
	-o-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-ms-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-moz-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-webkit-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
}

.small-graph-container {
	box-sizing: border-box;
	width: 1200px;
	height: 220px;
	padding: 20px 15px 15px 15px;
	margin: 15px auto 30px auto;
	border: 1px solid #ddd;
	background: #fff;
	background: linear-gradient(#f6f6f6 0, #fff 50px);
	background: -o-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -ms-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -moz-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -webkit-linear-gradient(#f6f6f6 0, #fff 50px);
	box-shadow: 0 3px 10px rgba(0,0,0,0.15);
	-o-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-ms-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-moz-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-webkit-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
}

.legend-container {
	box-sizing: border-box;
	width: 1200px;
	height: 450px;
	padding: 2px 1px 1px 1px;
	margin: 15px auto 30px auto;
	border: 0px;
	background: #fff;
	background: linear-gradient(#f6f6f6 0, #fff 50px);
	background: -o-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -ms-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -moz-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -webkit-linear-gradient(#f6f6f6 0, #fff 50px);
	box-shadow: 0 3px 10px rgba(0,0,0,0.15);
	-o-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-ms-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-moz-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-webkit-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
}


.demo-placeholder {
	width: 100%;
	height: 100%;
	font-size: 14px;
	line-height: 1.2em;
}
</style>
</head>
<body>
<pre>{{ .Title }}</pre>
<div id="export">
	<a href="/graph.json">json</a>
</div>
<div id="content">

	<div class="graph-container">
		<div id="datagraph" class="demo-placeholder"></div>
	</div>

	<div class="small-graph-container">
		<div id="clockgraph" class="demo-placeholder"></div>
	</div>

	<div class="small-graph-container">
		<div id="cpugraph" class="demo-placeholder"></div>
	</div>

	<div class="legend-container" style="height:60px;">
		<div id="overview" class="demo-placeholder"></div>
	</div>

	<p>The smaller plot is linked to the main plot, so it acts as an overview. Try dragging a selection on either plot, and watch the behavior of the other.</p>

</div>

<pre><b>Legend</b>
<dl>

<dt>gc.heapinuse  </dt><dd> heap in use after gc</dd>
<dt>scvg.inuse    </dt><dd> virtual memory considered in use by the scavenger</dd>
<dt>scvg.idle     </dt><dd> virtual memory considered unused by the scavenger</dd>
<dt>scvg.sys      </dt><dd> virtual memory requested from the operating system (should aproximate VSS)</dd>
<dt>scvg.released </dt><dd> virtual memory returned to the operating system by the scavenger</dd>
<dt>scvg.consumed </dt><dd> virtual memory in use (should roughly match process RSS)</dd>

<dt>STW sweep clock   </dt><dd>stop-the-world sweep clock time</dd>
<dt>con mas clock     </dt><dd>concurrent mark and scan clock time</dd>
<dt>STW mark clock    </dt><dd>stop-the-world mark clock time</dd>
<dt>STW sweep cpu     </dt><dd>stop-the-world sweep cpu time</dd>
<dt>con mas assist cpu</dt><dd>concurrent mark and scan - assist cpu time (GC performed in line with allocation)</dd>
<dt>con mas bg cpu    </dt><dd>concurrent mark and scan - background GC cpu time</dd>
<dt>con mas idle cpu  </dt><dd>concurrent mark and scan - idle GC cpu time</dd>
<dt>STW mark cpu      </dt><dd>stop-the-world mark cpu time</dd>
</dl>

</pre>
</body>
</html>
	`
)
