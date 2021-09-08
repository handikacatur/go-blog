package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Options struct {
	Image ImageOptions
}

type ImageOptions struct {
	Classes ImageClasses
	Caption string
}

type ImageClasses struct {
	WithBorder     string
	Stretched      string
	WithBackground string
}

type EditorJS struct {
	Blocks []EditorJSBlock `json:"blocks"`
}

type EditorJSBlock struct {
	Type string       `json:"type"`
	Data EditorJSData `json:"data"`
}

type EditorJSData struct {
	Text           string     `json:"text"`
	Level          int        `json:"level" `
	Style          string     `json:"style" `
	Items          []string   `json:"items" `
	File           FileData   `json:"file" `
	Caption        string     `json:"caption"`
	WithBorder     bool       `json:"withBorder"`
	Stretched      bool       `json:"stretched"`
	WithBackground bool       `json:"withBackground"`
	HTML           string     `json:"html"`
	Content        [][]string `json:"content"`
	Alignment      string     `json:"alignment"`
	Url            string     `json:"url"`
}

type FileData struct {
	URL string `json:"url"`
}

func ParseEditorJSON(editorJS string) EditorJS {
	var result EditorJS

	err := json.Unmarshal([]byte(editorJS), &result)
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func HTML(input string, options ...Options) string {
	var markdownOptions Options

	if len(options) > 0 {
		markdownOptions = options[0]
	}

	var result []string
	editorJSAST := ParseEditorJSON(input)

	for _, el := range editorJSAST.Blocks {

		data := el.Data

		switch el.Type {

		case "header":
			result = append(result, generateHTMLHeader(data))

		case "paragraph":
			result = append(result, generateHTMLParagraph(el.Data))

		case "list":
			result = append(result, generateMDList(data))

		case "simple-image":
			result = append(result, generateHTMLImage(data, markdownOptions))

		case "rawTool":
			result = append(result, data.HTML)

		case "delimiter":
			result = append(result, "---")

		case "table":
			result = append(result, generateMDTable(data))

		case "caption":
			result = append(result, generateMDCaption(data))

		case "image":
			result = append(result, generateSimpleImage(data))

		default:
			log.Fatal("Unknown data type: " + el.Type)
		}

	}

	return strings.Join(result[:], "\n\n")
}

func generateHTMLHeader(el EditorJSData) string {
	level := strconv.Itoa(el.Level)
	return fmt.Sprintf("<h%s>%s</h%s>", level, el.Text, level)
}

func generateHTMLParagraph(el EditorJSData) string {
	return fmt.Sprintf("<p>%s</p>", el.Text)
}

// func generateHTMLList(el EditorJSData) string {
// 	var result []string

// 	if el.Style == "unordered" {
// 		result = append(result, "<ul>")

// 		for _, el := range el.Items {
// 			result = append(result, "  <li>"+el+"</li>")
// 		}

// 		result = append(result, "</ul>")
// 	} else {
// 		result = append(result, "<ol>")

// 		for _, el := range el.Items {
// 			result = append(result, "  <li>"+el+"</li>")
// 		}

// 		result = append(result, "</ol>")
// 	}

// 	return strings.Join(result[:], "\n")
// }

func generateHTMLImage(el EditorJSData, options Options) string {
	classes := options.Image.Classes
	withBorder := classes.WithBorder
	stretched := classes.Stretched
	withBackground := classes.WithBackground

	if withBorder == "" && el.WithBorder {
		withBorder = "editorjs-with-border"
	}

	if stretched == "" && el.Stretched {
		stretched = "editorjs-stretched"
	}

	if withBackground == "" && el.WithBackground {
		withBackground = "editorjs-withBackground"
	}

	return fmt.Sprintf(`<img src="%s" alt="%s" class="%s %s %s" />`, el.File.URL, options.Image.Caption, withBorder, stretched, withBackground)
}

func generateSimpleImage(el EditorJSData) string {
	return fmt.Sprintf(`<img src="%s" alt="%s" class="img-fluid"`, el.Url, el.Caption)
}

func generateMDList(el EditorJSData) string {
	var result []string

	if el.Style == "unordered" {
		for _, el := range el.Items {
			result = append(result, "- "+el)
		}
	} else {
		for i, el := range el.Items {
			n := strconv.Itoa(i+1) + "."
			result = append(result, fmt.Sprintf("%s %s", n, el))
		}
	}

	return strings.Join(result[:], "\n")
}

func generateMDTable(el EditorJSData) string {
	var result []string

	for _, cell := range el.Content {
		row := strings.Join(cell, " | ")
		result = append(result, fmt.Sprintf("| %s |", row))
	}

	return strings.Join(result, "\n")
}

func generateMDCaption(el EditorJSData) string {
	return fmt.Sprintf("> %s", el.Text)
}
