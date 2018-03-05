var index = {
    init: function() {
        // Init
        asticode.loader.init();
        asticode.notifier.init();

        // Wait for astilectron to be ready
        document.addEventListener('astilectron-ready', function() {
            // Reset
            index.reset();
        })
    },
    reset: function() {
        document.getElementById("content").innerHTML = '<button id="btn-start" class="btn-sm btn-success">Start</button>';
        document.getElementById("btn-start").onclick = function() {
            astilectron.showOpenDialog({message: "Select files", properties: ['openFile', 'multiSelections'], title: "Select files"}, function(paths) {
                if (typeof paths !== "undefined" && paths.length > 0) {
                    index.selectAction(paths);
                }
            })
        };
    },
    selectAction: function(inputPaths) {
        document.getElementById("content").innerHTML = `<button id="btn-action-bitrate" class="btn-sm btn-success btn-margin">Bitrate</button>
        <button id="btn-action-psnr" class="btn-sm btn-success btn-margin">PSNR</button>
        <div id="charts"></div>`;
        document.getElementById("btn-action-bitrate").onclick = function() {
            index.visualize("bitrate", {input_paths: inputPaths});
        };
        document.getElementById("btn-action-psnr").onclick = function() {
            astilectron.showOpenDialog({message: "Select source", properties: ['openFile'], title: "Select source"}, function(paths) {
                if (typeof paths !== "undefined" && paths.length > 0) {
                    index.visualize("psnr", {input_paths: inputPaths, source_path: paths[0]});
                }
            })
        };
    },
    visualize: function(action, payload) {
        asticode.loader.show();
        astilectron.sendMessage({"name": "visualize." + action, payload: payload}, function(message) {
            // Hide loader
            asticode.loader.hide();

            // Hide button
            document.getElementById("btn-action-"+action).style.display = "none";
            document.getElementById("content").style.verticalAlign = "top";

            // Process error
            if (message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }

            // No charts
            if (message.payload.length === 0) {
                return
            }

            // Loop through charts
            for (let i = 0; i < message.payload.length; i++) {
                let canvas = document.createElement("canvas");
                document.getElementById("charts").append(canvas);
                new Chart(canvas, message.payload[i]);
            }
        })
    }
};