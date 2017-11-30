This is a set of video tools available through a nice UI.

# Prerequisites (for now)

The following must be available on your machine:
- ffprobe

# Install the project

This project has to be installed similarly to the [demo of astilectron](https://github.com/asticode/go-astilectron-demo).

## Step 1: install the project

Run the following command:

    $ go get -u github.com/asticode/go-astivid/...
    $ rm $GOPATH/src/github.com/asticode/go-astivid/bind.go

## Step 2: install the bundler

Run the following command:

    $ go get -u github.com/asticode/go-astilectron-bundler/...
    
## Step 3: bundle the app for your current environment

Run the following commands:

    $ cd $GOPATH/src/github.com/asticode/go-astivid
    $ astilectron-bundler -v
    
## Step 4: test the app

The result is in the `output/<your os>-<your arch>` folder and is waiting for you to test it!

## Step 5: bundle the app for more environments

To bundle the app for more environments, add an `environments` key to the bundler configuration (`bundler.json`):

```json
"environments": [
  {"arch": "amd64", "os": "linux"},
  {"arch": "386", "os": "windows"}
]
```

and repeat **step 3**.