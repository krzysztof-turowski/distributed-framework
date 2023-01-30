# distributed-framework
A collection of algorithms for _Distributed Systems_ course (winter semesters 2020/21, 2021/22, 2022/23) at [Jagiellonian University](https://uj.edu.pl), [Theoretical Computer Science Department](https://tcs.uj.edu.pl).

## Algorithms

### Leader election

#### Synchronized directed ring
1. Chang-Roberts algorithm
2. Itai-Rodeh algorithm
3. Dolev-Klawe-Rodeh algorithms A and B
4. Peterson algorithm

#### Synchronized undirected ring
1. Hirschberg-Sinclair algorithm
2. Korach-Rotem-Santoro ProbAsFar algorithm

#### Synchronized undirected mesh
1. Peterson algorithm

#### Synchronized oriented hypercube
1. Hyperelect algorithm

#### Synchronized oriented clique
1. Humblet algorithm

#### Synchronized undirected graph
1. Yo-Yo algorithm

#### Asynchronous directed ring
1. Itai-Rodeh algorithm

#### Asynchronous undirected ring
1. Stages with feedback (Korach-Rotem-Santoro) algorithm

### Consensus

#### Synchronized network
1. Single Bit (Garay-Berman) algorithm
2. Phase King (Garay-Berman-Perry) algorithm
3. Ben-Or randomized algorithm

### Graph algorithms

#### Synchronized minimum spanning tree
1. Gallager-Humblet-Spira algorithm

#### Synchronized maximal independent set
1. Luby algorithm

## Running

Run example:
```bash
  go run src/example/synchronized.go 5
```
