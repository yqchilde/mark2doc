package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed assets/mdtht.min.css assets/mdtht.min.js assets/highlight.min.js
var assets embed.FS

func main() {
	// 定义命令行参数
	inputFile := flag.String("i", "", "输入Markdown文件路径 (必填)")
	outputFile := flag.String("o", "", "输出HTML文件路径 (必填)")
	customTitle := flag.String("title", "", "自定义HTML文档标题 (可选，默认使用文件名)")
	flag.Parse()

	// 检查必填参数
	if *inputFile == "" || *outputFile == "" {
		fmt.Println("错误: 必须同时指定输入和输出文件")
		fmt.Println("用法: mark2doc -i <markdown文件> -o <html文件> [-title <自定义标题>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 读取Markdown文件
	mdData, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("读取文件错误: %v\n", err)
		os.Exit(1)
	}

	// 获取文件标题作为页面标题
	var title string
	if *customTitle != "" {
		// 使用自定义标题
		title = *customTitle
	} else {
		// 使用文件名作为标题
		title = filepath.Base(*inputFile)
		title = strings.TrimSuffix(title, filepath.Ext(title))
	}

	// 创建markdown解析器
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // 允许原始HTML，请注意安全风险
		),
	)

	// 渲染Markdown内容为HTML
	var buf bytes.Buffer
	if err := md.Convert(mdData, &buf); err != nil {
		fmt.Printf("渲染Markdown内容错误: %v\n", err)
		os.Exit(1)
	}
	contentHTML := buf.Bytes()

	// 读取嵌入的资源文件
	mdthtCss, _ := assets.ReadFile("assets/mdtht.min.css")
	mdthtJs, _ := assets.ReadFile("assets/mdtht.min.js")
	highlightJs, _ := assets.ReadFile("assets/highlight.min.js")

	// 构建完整的HTML文档
	htmlTemplate := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>%s</style>
    <script>%s</script>
	<script>%s</script>
    <script>hljs.highlightAll();</script>
</head>
<body>
    %s
</body>
</html>`

	// 格式化HTML文档
	htmlData := []byte(fmt.Sprintf(
		htmlTemplate,
		title,
		mdthtCss,
		mdthtJs,
		highlightJs,
		contentHTML,
	))

	// 写入HTML文件
	err = os.WriteFile(*outputFile, htmlData, 0644)
	if err != nil {
		fmt.Printf("生成HTML文件错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功将 %s 转换为 %s\n", *inputFile, *outputFile)
}
