package dialog

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PickFolder abre el selector de carpetas nativo del SO.
// Retorna (path, true) si el usuario selecciono una carpeta,
// (\"\", false) si cancelo, y error si fallo el selector.
func PickFolder(title string) (string, bool, error) {
	if title == "" {
		title = "Selecciona tu Vault"
	}
	switch runtime.GOOS {
	case "windows":
		return pickFolderWindows(title)
	case "darwin":
		return pickFolderDarwin(title)
	default:
		return pickFolderLinux(title)
	}
}

func pickFolderWindows(title string) (string, bool, error) {
	// PowerShell FolderBrowserDialog — nativo, no pide escribir ruta
	ps := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$dlg = New-Object System.Windows.Forms.FolderBrowserDialog
$dlg.Description = "%s"
$dlg.ShowNewFolderButton = $true
if($dlg.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK){
  Write-Output $dlg.SelectedPath
}
`, escapePS(title))

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps)
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		// intenta con pwsh si powershell no existe
		cmd2 := exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", ps)
		cmd2.Stdout = &out
		cmd2.Stderr = &errBuf
		if err2 := cmd2.Run(); err2 != nil {
			return "", false, fmt.Errorf("selector carpetas fallo: %v (%s)", err, strings.TrimSpace(errBuf.String()))
		}
	}
	sel := strings.TrimSpace(out.String())
	if sel == "" {
		return "", false, nil
	}
	// validar que no haya path traversal raro
	if _, err := os.Stat(sel); err != nil && !os.IsNotExist(err) {
		return "", false, fmt.Errorf("ruta invalida: %w", err)
	}
	return sel, true, nil
}

func pickFolderDarwin(title string) (string, bool, error) {
	script := fmt.Sprintf(`set d to choose folder with prompt "%s"
POSIX path of d`, escapeApple(title))
	cmd := exec.Command("osascript", "-e", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if strings.Contains(out.String(), "User canceled") {
			return "", false, nil
		}
		return "", false, fmt.Errorf("osascript fallo: %v (%s)", err, strings.TrimSpace(out.String()))
	}
	sel := strings.TrimSpace(out.String())
	if sel == "" {
		return "", false, nil
	}
	return sel, true, nil
}

func pickFolderLinux(title string) (string, bool, error) {
	// intenta zenity, luego kdialog, luego yad
	for _, bin := range []string{"zenity", "kdialog", "yad"} {
		if _, err := exec.LookPath(bin); err != nil {
			continue
		}
		var cmd *exec.Cmd
		switch bin {
		case "zenity":
			cmd = exec.Command(bin, "--file-selection", "--directory", "--title="+title)
		case "kdialog":
			cmd = exec.Command(bin, "--getexistingdirectory", ".", "--title", title)
		case "yad":
			cmd = exec.Command(bin, "--file-selection", "--directory", "--title="+title)
		}
		var out bytes.Buffer
		cmd.Stdout = &out
		_ = cmd.Run()
		sel := strings.TrimSpace(out.String())
		if sel != "" {
			return sel, true, nil
		}
		// si no hay salida, consideramos cancelado sin error
		return "", false, nil
	}
	return "", false, fmt.Errorf("no hay selector nativo (instala zenity/kdialog/yad)")
}

func escapePS(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func escapeApple(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
