# distributed-framework
A collection of algorithms for _Distributed Systems_ course (winter semesters 2020/21, 2021/22, 2022/23, 2023/24, 2024/25) at [Jagiellonian University](https://uj.edu.pl), [Theoretical Computer Science Department](https://tcs.uj.edu.pl).

## Algorithms

### Leader election

#### Synchronized directed ring
1. Chang-Roberts algorithm
1. Peterson (Dolev-Klawe-Rodeh A) algorithm
1. Improved Peterson algorithm
1. Dolev-Klawe-Rodeh algorithm B
1. Simplified Itai-Rodeh algorithm
1. Itai-Rodeh algorithm
1. Higham-Przytycka algorithm

#### Synchronized undirected ring
1. Franklin algorithm
1. Hirschberg-Sinclair algorithm
1. ProbAsFar (Korach-Rotem-Santoro) algorithm
1. Higham-Przytycka algorithm

#### Synchronized undirected mesh
1. Peterson algorithm

#### Synchronized oriented hypercube
1. Hyperelect algorithm

#### Synchronized torus
1. Mans orientation algorithm (prelude to Peterson leader election)

#### Synchronized oriented clique
1. Humblet algorithm

#### Synchronized undirected graph
1. Yo-Yo algorithm
1. Casteigts-Métivier-Robson-Zemmari algorithm

#### Asynchronous directed ring
1. Chang-Roberts algorithm
1. Peterson (Dolev-Klawe-Rodeh A) algorithm
1. Improved Peterson algorithm
1. Dolev-Klawe-Rodeh algorithm B
1. Itai-Rodeh algorithm
1. Higham-Przytycka algorithm

#### Asynchronous undirected ring
1. Franklin algorithm
1. Hirschberg-Sinclair algorithm
1. Stages with feedback (Korach-Rotem-Santoro) algorithm
1. Probabilistic Franklin algorithm
1. Itai-Rodeh algorithm
1. Higham-Przytycka algorithm

#### Asynchronous clique
1. Korach-Moran-Zaks algorithm
2. Afek-Gafni B algorithm

#### Asynchronous oriented clique
1. Loui-Matsushita-West algorithm

### Graph orientation

#### Asynchronous ring
1. Syrotiuk-Pachl algorithm

### Order estimation

#### Asynchronous undirected ring
1. Itai-Rodeh algorithm

### Consensus

#### Byzantine problem

#### Synchronized network
1. Single Bit (Garay-Berman) algorithm
1. Phase King (Garay-Berman-Perry) algorithm
1. Ben-Or randomized algorithm
1. Chor-Coan algorithm

### Graph algorithms

#### Synchronized minimum spanning tree
1. Gallager-Humblet-Spira algorithm

#### Synchronized maximal independent set
1. Luby algorithm
2. Métivier-Robson-Saheb-Djahromi-Zemmari algorithm (C)

#### Synchronized dominating set
1. Local Randomized Greedy (Jia-Rajaraman-Suel) algorithm
1. Kuhn-Wattenhoffer algorithm

## Running

Run example:
```bash
  go run src/example/synchronized.go 5
```
