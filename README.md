# distributed-framework
A collection of algorithms for _Distributed Systems_ course (winter semester 2020/21) at [Jagiellonian University](https://uj.edu.pl), [Theoretical Computer Science Department](https://tcs.uj.edu.pl).

## Algorithms

### Leader election

#### Synchronized directed ring
1. Chang-Roberts algorithm
2. Itai-Rodeh algorithm
3. Dolev-Klawe-Rodeh algorithms A and B

#### Synchronized undirected mesh
1. Peterson algorithm

#### Synchronized undirected graph
1. YO-YO algorithm

### Graph algorithms

#### Minimum spanning tree
1. Synchronized Gallager, Humblet, and Spira algorithm
#### Maximal independent set
1. Luby algorithm

## Running

Run example:
```bash
  go run src/example/synchronized.go 5
```
