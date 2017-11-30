var index = {
    analyze: function(paths) {
        asticode.loader.show();
        astilectron.sendMessage({"name": "get.frames", payload: paths}, function(message) {
            // Hide loader
            asticode.loader.hide();

            // Process error
            if (message.name === "error") {
                asticode.notifier.error(message.payload)
                return
            }

            // No charts
            if (message.payload.length === 0) {
                return
            }

            // Loop through charts
            for (var i = 0; i < message.payload.length; i++) {
                var canvas = document.createElement("canvas");
                canvas.id = "chart-" + index.countCharts;
                document.getElementById("charts").append(canvas);
                new Chart(canvas, message.payload[i]);
                index.countCharts++
            }
        })
    },
    init: function() {
        // Init
        index.countCharts = 0;
        asticode.loader.init();
        asticode.notifier.init();

        // Wait for astilectron to be ready
        document.addEventListener('astilectron-ready', function() {
            // Handle buttons
            index.handleButtons();
        })
    },
    handleButtons: function() {
        // Analyze button
        document.getElementById("btn-analyze").onclick = function() {
            astilectron.showOpenDialog({properties: ['openFile', 'multiSelections']}, function(paths) {
                index.analyze(paths);
            })
        }
    }
};