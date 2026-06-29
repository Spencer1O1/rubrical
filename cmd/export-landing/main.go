package main

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"rubrical/internal/web/components"
	"rubrical/internal/web/pages"
)

func main() {
	if err := os.MkdirAll("public/static/css", 0o755); err != nil {
		panic(err)
	}

	index, err := os.Create("public/index.html")
	if err != nil {
		panic(err)
	}
	defer index.Close()

	if err := pages.Landing(components.MarketingNav{}).Render(context.Background(), index); err != nil {
		panic(err)
	}

	if err := copyFile("static/css/app.css", "public/static/css/app.css"); err != nil {
		panic(err)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
