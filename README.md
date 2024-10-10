# go-cubic

A simple golang application for creating NxNxN cubes, applying algorithms, and generating an HTML output of the cube's state.

## Features

- Create a cube of any size $N \times N$.
- Parse official WCA notation, including:
    - Commutators, conjugates, and parenthesis groups with factor suffixes.
- Apply moves to the cube.
- Retrieve the current state of the cube's faces.
- Generate an HTML file to display the cube's state.

## Installation

For using go-cubic as is, you can simply clone the repo and run it.

1. Clone the repository.
```bash
git clone https://github.com/larssont/go-cubic
```

2. Install the dependencies.
```bash
go mod tidy
```

3. Fire it up.
```bash
go run cmd/main.go
```


## Usage

Creating a 4x4 cube:

```go
c := cube.NewCube(4)
```

Parse notation into a group and print it:
```go
input := "(R L2 U')2"

group, _ := cube.ParseNotation(input)

group.Print()
```

Apply moves from a group onto a cube.
> [!TIP]
> Official WCA notation, as well as commutators and conjugates, is supported. Please see the [WCA Regulations](https://www.worldcubeassociation.org/regulations/#article-12-notation) for details.
```go
c := cube.NewCube(3)
input := "[L U : [S' , L2']]"

group, _ := cube.ParseNotation(input)
moves, _ := group.Expand() // Expand to get a slice of all moves 

c.ExecuteMoves(moves)
```

The faces of the cube can also be retrieved:
```go
c := cube.NewCube(7)

faces := c.Faces()
```

The `CubeFaces` struct is defined as follows:
```go
type CubeFaces struct {
	Up    []rune // A slice of runes associated with a color. w/y/g/b/r/o
	Left  []rune
	Front []rune
	Right []rune
	Back  []rune
	Down  []rune
}
```

Feel free to take a look at [main.go](cmd/main.go) for an example of creating and displaying a cube on a website with the help of go's html/template, css transformations, and some javascript.

```go
input := "U2 R2 B L2 U2 D2 F' U2 F B2 L' B' D F U L' B' D R2 Fw2 D L Rw2 Fw2 L D2 F2 L' U' R' F Fw' D' R2 D B2 Rw' Uw2 R Rw' D Rw2 Uw'"

group, _ := cube.ParseNotation(input)
moves, _ := group.Expand()
c := cube.NewCube(4)
c.ExecuteMoves(moves...)

GenerateHTML(c, "cube.html")
```

![Rotating 4x4 Cube](/assets/cube-4x4.gif)


## License

This project is licensed under the MPL 2.0 License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

Use this space to list resources you find helpful and would like to give credit to. I've included a few of my favorites to kick things off!

* [Choose an Open Source License](https://choosealicense.com)
* [Ana Tudor](https://www.youtube.com/watch?v=xvxXgcvUY_w)