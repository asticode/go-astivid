const colors = ["blue","green","red","yellow","purple","orange"];
let index = {
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
        document.getElementById("content").style.verticalAlign = "middle";
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
        let pathsWithColor = {};
        for (let idx = 0; idx < colors.length; idx++) {
            if (typeof inputPaths[idx] !== "undefined") {
                pathsWithColor[colors[idx]] = inputPaths[idx];
            }
        }
        document.getElementById("content").innerHTML = `<button id="btn-action-bitrate" class="btn-sm btn-success btn-margin">Bitrate</button>
        <button id="btn-action-psnr" class="btn-sm btn-success btn-margin">PSNR</button>
        <button id="btn-action-reset" class="btn-sm btn-success btn-margin">Reset</button>
        <div id="charts"></div>`;
        document.getElementById("btn-action-bitrate").onclick = function() {
            index.visualize("bitrate", {input_paths: pathsWithColor});
        };
        document.getElementById("btn-action-psnr").onclick = function() {
            astilectron.showOpenDialog({message: "Select source", properties: ['openFile'], title: "Select source"}, function(paths) {
                if (typeof paths !== "undefined" && paths.length > 0) {
                    index.visualize("psnr", {input_paths: pathsWithColor, source_path: paths[0]});
                }
            })
        };
        document.getElementById("btn-action-reset").onclick = function() {
            index.reset();
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
                let wrapper = document.createElement("div");
                document.getElementById("charts").append(wrapper);
                let header = document.createElement("div");
                header.style.height = "20px";
                header.style.textAlign = "right";
                wrapper.append(header);
                let download = document.createElement("a");
                header.append(download);
                let canvas = document.createElement("canvas");
                wrapper.append(canvas);
                let payload = message.payload[i];
                let c;
                let done = false;
                payload.options.animation = {onComplete: function() {
                    if (!done) {
                        download.outerHTML = "<a href='" + c.toBase64Image() + "' style='color: inherit' download><i class='fa fa-download'></i></a>";
                        done = true;
                    }
                }};
                c = new Chart(canvas, payload);
            }
        })
    }
};