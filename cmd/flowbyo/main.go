package main

import (
	"flag"
	"fmt"
	"os"

	"goflowyourself/pkg/diagram"
)

func main() {
	loadCmd := flag.NewFlagSet("load", flag.ExitOnError)
	inputFile := loadCmd.String("input", "", "Input YAML file")
	outputSVG := loadCmd.String("output", "", "Output SVG file")
	outputHTML := loadCmd.String("html", "", "Output HTML file")
	validate := loadCmd.Bool("validate", false, "Run validation only")
	save := loadCmd.String("save", "", "Save back to YAML file")

	fromSchemaCmd := flag.NewFlagSet("from-schema", flag.ExitOnError)
	fromSchemaInput := fromSchemaCmd.String("input", "", "Input text/markdown schema file")
	fromSchemaOutput := fromSchemaCmd.String("output", "", "Output SVG file")
	fromSchemaHTML := fromSchemaCmd.String("html", "", "Output HTML file")
	fromSchemaSave := fromSchemaCmd.String("save", "", "Save graph-based YAML file")
	fromSchemaName := fromSchemaCmd.String("name", "Text Schema Diagram", "Diagram name")
	fromSchemaType := fromSchemaCmd.String("diagram-type", "block-diagram", "Diagram type")
	fromSchemaVersion := fromSchemaCmd.String("version", "1", "Diagram version")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "load":
		loadCmd.Parse(os.Args[2:])
		handleLoad(*inputFile, *outputSVG, *outputHTML, *validate, *save)
	case "from-schema":
		fromSchemaCmd.Parse(os.Args[2:])
		handleFromSchema(*fromSchemaInput, *fromSchemaOutput, *fromSchemaHTML, *fromSchemaSave, *fromSchemaName, *fromSchemaType, *fromSchemaVersion)
	case "version":
		fmt.Println("goflowyourself v0.1.0 - Nested Matrix Diagram MVP")
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func handleFromSchema(input, outputSVG, outputHTML, saveFile, name, diagramType, version string) {
	if input == "" {
		fmt.Println("Error: --input is required")
		os.Exit(1)
	}
	if outputSVG == "" && outputHTML == "" && saveFile == "" {
		fmt.Println("Error: provide at least one output target via --output, --html, or --save")
		os.Exit(1)
	}

	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Printf("Error reading schema: %v\n", err)
		os.Exit(1)
	}

	d, err := diagram.ParseTextSchemaToDiagram(name, diagramType, version, string(data))
	if err != nil {
		fmt.Printf("Error parsing text schema: %v\n", err)
		os.Exit(1)
	}

	model := diagram.BuildRuntimeModel(d)
	diagram.Validate(model)
	fmt.Printf("✓ Parsed %d elements from text schema\n", len(model.Elements))
	model.PrintValidation()

	if model.HasErrors() {
		fmt.Println("Validation failed. Aborting.")
		os.Exit(1)
	}

	if err := renderOutputs(model, outputSVG, outputHTML); err != nil {
		fmt.Printf("Error rendering output: %v\n", err)
		os.Exit(1)
	}

	if saveFile != "" {
		saveDiagram := &diagram.Diagram{
			Meta:  d.Meta,
			Graph: model.Graph,
			View:  d.View,
		}
		if err := diagram.SaveYAML(saveDiagram, saveFile); err != nil {
			fmt.Printf("Error saving YAML: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Graph YAML saved: %s\n", saveFile)
	}
}

func handleLoad(input, outputSVG, outputHTML string, validateOnly bool, saveFile string) {
	if input == "" {
		fmt.Println("Error: --input is required")
		os.Exit(1)
	}

	// Parse YAML
	fmt.Printf("Loading diagram from: %s\n", input)
	d, err := diagram.ParseYAML(input)
	if err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Loaded: %s (%s)\n", d.Meta.Name, d.Meta.DiagramType)

	// Build runtime model
	model := diagram.BuildRuntimeModel(d)

	// Validate
	diagram.Validate(model)
	fmt.Printf("✓ Found %d elements\n", len(model.Elements))
	model.PrintValidation()

	if model.HasErrors() {
		fmt.Println("Validation failed. Aborting.")
		os.Exit(1)
	}

	if validateOnly {
		fmt.Println("✓ Validation passed")
		return
	}

	// Render to SVG if requested
	if err := renderOutputs(model, outputSVG, outputHTML); err != nil {
		fmt.Printf("Error rendering output: %v\n", err)
		os.Exit(1)
	}

	// Save back to YAML if requested
	if saveFile != "" {
		fmt.Printf("Saving to: %s\n", saveFile)
		saveDiagram := &diagram.Diagram{
			Meta:  d.Meta,
			Graph: model.Graph,
			View:  d.View,
		}
		err := diagram.SaveYAML(saveDiagram, saveFile)
		if err != nil {
			fmt.Printf("Error saving YAML: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Saved")
	}

	fmt.Println("✓ Done")
}

func renderOutputs(model *diagram.RuntimeModel, outputSVG, outputHTML string) error {
	renderer := diagram.NewSVGRenderer()

	if outputSVG != "" {
		fmt.Printf("Rendering to SVG: %s\n", outputSVG)
		if err := renderer.RenderToSVG(model, outputSVG); err != nil {
			return err
		}
		fmt.Println("✓ SVG generated")
	}

	if outputHTML != "" {
		fmt.Printf("Rendering to HTML: %s\n", outputHTML)
		if err := renderer.RenderToHTML(model, outputHTML); err != nil {
			return err
		}
		fmt.Println("✓ HTML generated")
	}

	return nil
}

func printUsage() {
	fmt.Println(`goflowyourself - Nested Matrix Diagram Tool

Usage:
	goflowyourself load --input <file.yaml> [--output <file.svg>] [--html <file.html>] [--save <file.yaml>] [--validate]
	goflowyourself from-schema --input <schema.txt> [--output <file.svg>] [--html <file.html>] [--save <file.yaml>]
	goflowyourself version
	goflowyourself help

Commands:
	load        Parse graph-based YAML, validate, and render a diagram
	from-schema Parse ASCII/markdown schema text and generate diagram outputs
  version    Show version
  help       Show this help message

Flags:
  --input      Input YAML diagram file (required)
	--output     Output SVG file (optional)
	--html       Output HTML file (optional)
  --validate   Validate only, don't render (optional)
  --save       Save back to YAML file (optional)

Examples:
  goflowyourself load --input diagram.yaml --output diagram.svg
	goflowyourself load --input diagram.yaml --html diagram.html
	goflowyourself load --input diagram.yaml --validate
	goflowyourself from-schema --input schema.txt --output diagram.svg --html diagram.html --save diagram.yaml`)
}
