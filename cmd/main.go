package main

import (
	"fmt"
	"go-cubic/pkg/cube"
	"html/template"
	"os"
	"path"
)

var (
	tmpl *template.Template
)

const (
	outPath = "out/"
)

func GenerateHTML(c *cube.Cube, filename string) error {
	file, err := os.Create(path.Join(outPath, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		Faces     *cube.CubeFaces
		Dimension int
	}{c.Faces(), c.Dimension()}

	return tmpl.ExecuteTemplate(file, "cube", data)
}

func init() {
	var err error
	tmpl, err = template.ParseFiles("ui/html/cube.tmpl")
	if err != nil {
		panic("Failed to parse template: " + err.Error())
	}
}

func main() {
	input := "U2 R2 B L2 U2 D2 F' U2 F B2 L' B' D F U L' B' D R2 Fw2 D L Rw2 Fw2 L D2 F2 L' U' R' F Fw' D' R2 D B2 Rw' Uw2 R Rw' D Rw2 Uw'"

	group, err := cube.ParseNotation(input)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	group.Print()

	moves, _ := group.Expand()

	cube := cube.NewCube(4).ExecuteMoves(moves...)

	GenerateHTML(cube, "cube.html")
}
