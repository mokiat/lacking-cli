package macos

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	_ "embed"
	_ "image/png"

	"github.com/jackmordaunt/icns/v2"
	"github.com/mokiat/lacking-cli/internal/distribution"
	"github.com/urfave/cli"
)

const (
	platform = "macos"
	arch     = "amd64"
)

var (
	//go:embed templates/info.tmpl
	infoTemplate string
)

func Package(ctx *cli.Context) error {
	projectFolderLoc := ctx.Args().First()
	appFileLoc := filepath.Join(projectFolderLoc, "app.yml")

	desc, err := distribution.LoadAppDescriptor(appFileLoc)
	if err != nil {
		return fmt.Errorf("error loading app descriptor: %w", err)
	}

	distFolderLoc, err := distribution.CreateBlankDistFolder(projectFolderLoc, platform)
	if err != nil {
		return fmt.Errorf("error creating blank dist folder: %w", err)
	}

	appFolderLoc, err := createAppFolder(distFolderLoc, desc)
	if err != nil {
		return fmt.Errorf("error creating app folder: %w", err)
	}

	contentsFolderLoc, err := createContentsFolder(appFolderLoc)
	if err != nil {
		return fmt.Errorf("error creating contents folder: %w", err)
	}

	resourcesFolderLoc, err := createResourcesFolder(contentsFolderLoc)
	if err != nil {
		return fmt.Errorf("error creating resources folder: %w", err)
	}

	if err := createInfoFile(contentsFolderLoc, desc); err != nil {
		return fmt.Errorf("error creating info file: %w", err)
	}

	if err := createIconFile(projectFolderLoc, resourcesFolderLoc, desc); err != nil {
		return fmt.Errorf("error creating icon file: %w", err)
	}

	if err := createBinaryFile(projectFolderLoc, contentsFolderLoc, desc); err != nil {
		return fmt.Errorf("error creating binary file: %w", err)
	}

	return nil
}

func createAppFolder(distFolderLoc string, app *distribution.Application) (string, error) {
	appFolderLoc := filepath.Join(distFolderLoc, fmt.Sprintf("%s.app", app.Name))

	if err := os.MkdirAll(appFolderLoc, 0o775); err != nil {
		return "", fmt.Errorf("error creating folder: %w", err)
	}

	return appFolderLoc, nil
}

func createContentsFolder(appFolderLoc string) (string, error) {
	contentsFolderLoc := filepath.Join(appFolderLoc, "Contents")

	if err := os.MkdirAll(contentsFolderLoc, 0o775); err != nil {
		return "", fmt.Errorf("error creating folder: %w", err)
	}

	return contentsFolderLoc, nil
}

func createResourcesFolder(contentsFolderLoc string) (string, error) {
	resourcesFolderLoc := filepath.Join(contentsFolderLoc, "Resources")

	if err := os.MkdirAll(resourcesFolderLoc, 0o775); err != nil {
		return "", fmt.Errorf("error creating folder: %w", err)
	}

	return resourcesFolderLoc, nil
}

func createBinaryFile(projectFolderLoc, contentsFolderLoc string, app *distribution.Application) error {
	binFolderLoc := filepath.Join(contentsFolderLoc, "MacOS")
	if err := os.MkdirAll(binFolderLoc, 0o775); err != nil {
		return fmt.Errorf("error creating MacOS folder: %w", err)
	}

	binFileLoc := filepath.Join(binFolderLoc, app.ID)

	cmd := exec.Command("go", "build", "-o", binFileLoc, filepath.FromSlash(app.MainDir))
	cmd.Dir = projectFolderLoc
	cmd.Env = cmd.Environ()
	cmd.Env = append(cmd.Env, "GOOS", platform)
	cmd.Env = append(cmd.Env, "GOARCH", arch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build error: %w", err)
	}
	return nil
}

func createInfoFile(contentsFolderLoc string, app *distribution.Application) error {
	tmpl, err := template.New("desktop").Parse(infoTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	templateData := map[string]any{
		"ID":         app.ID,
		"LongID":     app.LongID,
		"Name":       app.Name,
		"Copyright":  app.Copyright,
		"Executable": app.ID,
		"Version":    app.Version,
	}

	var content bytes.Buffer
	if err := tmpl.Execute(&content, templateData); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	infoFileLoc := filepath.Join(contentsFolderLoc, "Info.plist")
	if err := os.WriteFile(infoFileLoc, content.Bytes(), 0o666); err != nil {
		return fmt.Errorf("error writing control file: %w", err)
	}
	return nil
}

func createIconFile(projectFolderLoc, resourcesFolderLoc string, app *distribution.Application) error {
	srcIconFileLoc := filepath.Join(projectFolderLoc, filepath.FromSlash(app.IconFile))
	srcFile, err := os.Open(srcIconFileLoc)
	if err != nil {
		return fmt.Errorf("error opening icon file: %w", err)
	}
	defer srcFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("error decoding icon image: %w", err)
	}

	dstIconFileLoc := filepath.Join(resourcesFolderLoc, fmt.Sprintf("%s.icns", app.ID))
	dstFile, err := os.Create(dstIconFileLoc)
	if err != nil {
		return fmt.Errorf("error creating icon file: %w", err)
	}
	defer dstFile.Close()

	if err := icns.Encode(dstFile, img); err != nil {
		return fmt.Errorf("error encoding icns file: %w", err)
	}
	return nil
}
