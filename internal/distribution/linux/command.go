package linux

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/mokiat/lacking-cli/internal/distribution"
	"github.com/urfave/cli"
)

const (
	platform = "linux"
	arch     = "amd64"
)

var (
	//go:embed templates/control.tmpl
	controlTemplate string

	//go:embed templates/desktop.tmpl
	desktopTemplate string
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

	packageFolderLoc, err := createPackageFolder(distFolderLoc, desc)
	if err != nil {
		return fmt.Errorf("error creating package folder: %w", err)
	}

	if err := createControlFile(packageFolderLoc, desc); err != nil {
		return fmt.Errorf("error creating control file: %w", err)
	}

	if err := createBinaryFile(projectFolderLoc, packageFolderLoc, desc); err != nil {
		return fmt.Errorf("error creating binary file: %w", err)
	}

	if err := createDesktopFile(packageFolderLoc, desc); err != nil {
		return fmt.Errorf("error creating desktop file: %w", err)
	}

	if err := createIconFile(projectFolderLoc, packageFolderLoc, desc); err != nil {
		return fmt.Errorf("error creating icon file: %w", err)
	}
	// TODO: Create icon file

	if err := createDebFile(distFolderLoc, packageFolderLoc); err != nil {
		return fmt.Errorf("error creating deb file: %w", err)
	}

	return nil
}

func createPackageFolder(distFolderLoc string, app *distribution.Application) (string, error) {
	packageFolderLoc := filepath.Join(distFolderLoc, fmt.Sprintf("%s_%s-%d_%s",
		app.ID,
		app.Version,
		time.Now().Unix(),
		arch,
	))

	if err := os.MkdirAll(packageFolderLoc, 0o775); err != nil {
		return "", fmt.Errorf("error creating folder: %w", err)
	}

	return packageFolderLoc, nil
}

func createControlFile(packageFolderLoc string, app *distribution.Application) error {
	debianFolderLoc := filepath.Join(packageFolderLoc, "DEBIAN")
	if err := os.Mkdir(debianFolderLoc, 0o775); err != nil {
		return fmt.Errorf("error creating DEBIAN folder: %w", err)
	}

	tmpl, err := template.New("control").Parse(controlTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	templateData := map[string]any{
		"Package":      app.ID,
		"Version":      app.Version,
		"Architecture": arch,
		"Maintainer":   app.Contact,
		"Description":  app.Description,
	}

	var content bytes.Buffer
	if err := tmpl.Execute(&content, templateData); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	controlFileLoc := filepath.Join(debianFolderLoc, "control")
	if err := os.WriteFile(controlFileLoc, content.Bytes(), 0o666); err != nil {
		return fmt.Errorf("error writing control file: %w", err)
	}
	return nil
}

func createBinaryFile(projectFolderLoc, packageFolderLoc string, app *distribution.Application) error {
	binFolderLoc := filepath.Join(packageFolderLoc, "usr", "local", "bin")
	if err := os.MkdirAll(binFolderLoc, 0o775); err != nil {
		return fmt.Errorf("error creating /usr/local/bin folder: %w", err)
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

func createDesktopFile(packageFolderLoc string, app *distribution.Application) error {
	appsFolderLoc := filepath.Join(packageFolderLoc, "usr", "share", "applications")
	if err := os.MkdirAll(appsFolderLoc, 0o775); err != nil {
		return fmt.Errorf("error creating /usr/share/applications folder: %w", err)
	}

	tmpl, err := template.New("desktop").Parse(desktopTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	templateData := map[string]any{
		"Name":    app.Name,
		"Comment": app.Description,
		"Icon":    app.ID,
		"Exec":    app.ID,
	}

	var content bytes.Buffer
	if err := tmpl.Execute(&content, templateData); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	desktopFileName := fmt.Sprintf("%s.desktop", app.ID)
	desktopFileLoc := filepath.Join(appsFolderLoc, desktopFileName)
	if err := os.WriteFile(desktopFileLoc, content.Bytes(), 0o666); err != nil {
		return fmt.Errorf("error writing control file: %w", err)
	}
	return nil
}

func createIconFile(projectFolderLoc, packageFolderLoc string, app *distribution.Application) error {
	iconsFolderLoc := filepath.Join(packageFolderLoc, "usr", "share", "icons", "hicolor", "512x512", "apps")
	if err := os.MkdirAll(iconsFolderLoc, 0o775); err != nil {
		return fmt.Errorf("error creating /usr/share/icons/hicolor/512x512/apps folder: %w", err)
	}

	srcIconFileLoc := filepath.Join(projectFolderLoc, filepath.FromSlash(app.IconFile))
	srcFile, err := os.Open(srcIconFileLoc)
	if err != nil {
		return fmt.Errorf("error opening icon file: %w", err)
	}
	defer srcFile.Close()

	dstIconFileLoc := filepath.Join(iconsFolderLoc, fmt.Sprintf("%s%s", app.ID, filepath.Ext(srcIconFileLoc)))
	dstFile, err := os.Create(dstIconFileLoc)
	if err != nil {
		return fmt.Errorf("error creating icon file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copying icon file: %w", err)
	}
	return nil
}

func createDebFile(distFolderLoc, packageFolderLoc string) error {
	packageFolderName := filepath.Base(packageFolderLoc)

	cmd := exec.Command("dpkg-deb", "--build", "--root-owner-group", packageFolderName)
	cmd.Dir = distFolderLoc
	cmd.Env = cmd.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dpkg-deb error: %w", err)
	}
	return nil
}
