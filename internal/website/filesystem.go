package website

import (
	"io/fs"
	"os"
	"path/filepath"
)

func docsFileSystem() fs.FS {
	return sectionFileSystem("ASSET_DIR", "docs")
}

func mainFileSystem() fs.FS {
	return sectionFileSystem("MAIN_ASSET_DIR", "main")
}

func showcaseFileSystem() fs.FS {
	return sectionFileSystem("SHOWCASE_ASSET_DIR", "showcase")
}

func sectionFileSystem(envName, section string) fs.FS {
	if dir := os.Getenv(envName); dir != "" {
		return os.DirFS(dir)
	}
	for _, dir := range []string{
		filepath.Join("deploy", "website", section),
		section,
		filepath.Join("..", "..", "deploy", "website", section),
	} {
		if _, err := os.Stat(filepath.Join(dir, "templates")); err == nil {
			return os.DirFS(dir)
		}
	}
	return os.DirFS(filepath.Join("deploy", "website", section))
}

func mustSubFS(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
