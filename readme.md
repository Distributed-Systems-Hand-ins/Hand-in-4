  ## How to run our program:

Requires Go (see the version in `go.mod`). Run everything from the project root (log files are written to the current
directory). Use one terminal per process.


**1. Start at least three nodes**, each in its own terminal, each with a
unique id:


go run ./node -id 1
go run ./node -id 2
go run ./node -id 3

Start all three nodes within a few seconds of each other. They will locate each other.
Each node will now request access to the critical section. Only one node can have access at a time, so they will determine an order by comparing their Lamport clocks and node IDs. 