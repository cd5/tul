# Tul compiler

This reposity contains the Tul compiler front-end.
It enables users to invoke the Tul compiler running in the cloud.

## Installation and setup

Clone and compile the Tul front-end:

    git clone https://github.com/tul-project/tul.git 
    cd tul 
    ./make.sh 

Next, create an account:

    mkdir -p ~/.config/tul 
    echo email=YOUR-EMAIL >> ~/.config/tul/account 
    echo password=PASSWORD >> ~/.config/tul/account 

An account is automatically registered with the server upon first invocation of the compiler.

Next, download a Go source code supported by the Tulgo compiler. Then compile and run it:

    cd .. 
    curl -L -o bt.go http://github.com/tul‑project/benchmarks/raw/master/go/binary‑tree.go 
    ./tul/bin/tulgo bt.go -o bt 
    ./bt -n=16

The only architecture supported by the compiler is linux/386.
Running the compiled binary in 64-bit Linux requires 32-bit glibc library.
