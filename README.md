# BRB

## Files

1. comm.go contains the stop and wait algorithm used for emulation shared memory
2. ewma.go is the algorithm used to estimate timeout time in the stop and wait protocol. https://datatracker.ietf.org/doc/html/rfc6298 contains the formula used. 
3. peer.go contains all the structures regarding each peer which is used in comm.go
4. brb.go contains the self stablizing broadcast algorithm 
5. primitives.go contains the primitives required for brb.go
6. irc.go contains the code that is required for multi round broadcasting.  

## Execution

1. Firstly list the number of processes' host:port in input.txt
2. To run a process metioned in input.txt use the following command from inside the code directory "go run *.go input.txt <line-number-of-process-in-input-file - 1>"

### Example 

For the current input.txt, open 3 terminals for each process and run 
 1. go run *.go input.txt 0
 2. go run *.go input.txt 1
 3. go run *.go input.txt 2

