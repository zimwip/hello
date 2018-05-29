# Simple GO server project

- Use Raft protocol to manage cluster
- https on localhost for chrome use 
    chrome://flags/#allow-insecure-localhost
    
Clean Architecture example taken from 
[http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/](http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/)


Install
-------

Create a new directory, where to host this project

    mkdir -p $GOPATH:src/github.com/zimwip/

Check out the source

    cd $GOPATH:src/github.com/zimwip/
    git clone https://github.com/zimwip/simple-go-service

Setup the GOPATH to include this path

    cd simple-go-server
    export GOPATH=$GOPATH:`pwd`

Then build the project

    dep ensure
    ./go_build.sh

Create the SQLite structure

    sqlite3 /var/tmp/production.sqlite < setup.sql

Run the server

    ./bin/server-linux-amd64

Access the web endpoint at [http://localhost:8080/orders?userId=40&orderId=60](http://localhost:8080/orders?userId=40&orderId=60)

To run the tests, for each module, run

    cd src/infrastructure && go test
    cd src/interfaces && go test

Enjoy.

