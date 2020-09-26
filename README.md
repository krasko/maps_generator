# maps_generator
A simple program for generating sensed and unsensed maps for small numbers of edges.

Generates all rooted/sensed/unsensed maps with a specified list of vertex degrees. For each map, the program prints three permutations ("angles", "edge sides", "edge ends") that define the map. Also, map's genus and orientability are computed.

The output can be easily `grep`ped to enumerate maps on specific surfaces. 

### Running the program

Generate all rooted maps with two vertices of degree 3 and two leaves: 

`go run main.go 3 3 1 1`

Generate all unsensed maps with two vertices of degree 4:

`go run main.go -unsensed 4 4`

Generate all sensed maps with two vertices of degree 4:

`go run main.go -sensed 4 4`

### Filtering the output: enumerating by surface

Enumerate 3-regular unsensed maps with six vertices on the Klein bottle:

`go run main.go -unsensed 3 3 3 3 3 3 | grep "^2 -" | wc`

Enumerate 3-regular sensed maps with six vertices on the torus:

`go run main.go -sensed 3 3 3 3 3 3 | grep "^1 +" | wc`

The first of three values printed by `wc` is the number of maps.