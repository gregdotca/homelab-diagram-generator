package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gregdotca/homelab-diagram-generator/internal/config"
	"github.com/gregdotca/homelab-diagram-generator/internal/layout"
	"github.com/gregdotca/homelab-diagram-generator/internal/render"
)

func main() {
	input := flag.String("input", "homelab.yaml", "path to YAML config file")
	output := flag.String("output", "homelab.svg", "path to output SVG file")
	flag.Parse()

	cfg, err := config.Load(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	scene := layout.Compute(cfg)
	svg := render.SVG(scene)

	if err := os.WriteFile(*output, []byte(svg), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %s (%.0f x %.0f)\n", *output, scene.Width, scene.Height)
}
