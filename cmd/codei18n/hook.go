package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/studyzy/codei18n/internal/log"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "管理 Git Hooks",
}

var hookInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "安装 pre-commit hook",
	Run: func(cmd *cobra.Command, args []string) {
		runHookInstall()
	},
}

var hookUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "卸载 pre-commit hook",
	Run: func(cmd *cobra.Command, args []string) {
		runHookUninstall()
	},
}

func init() {
	rootCmd.AddCommand(hookCmd)
	hookCmd.AddCommand(hookInstallCmd)
	hookCmd.AddCommand(hookUninstallCmd)
}

func runHookInstall() {
	gitDir := ".git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		log.Fatal("当前目录不是 Git 仓库根目录")
	}

	hookPath := filepath.Join(gitDir, "hooks", "pre-commit")

	// Create hook content
	// Pre-commit workflow:
	// 1. Update mappings for staged files (adds new Chinese comments)
	// 2. Translate missing English translations
	// 3. Convert comments to English
	// 4. Re-stage modified files
	hookContent := `#!/bin/sh
# CodeI18n Pre-commit Hook
# Converts comments to English before committing

echo "CodeI18n: Checking for non-English comments..."

# Get staged Go files
FILES=$(git diff --cached --name-only --diff-filter=ACM | grep "\.go$")

if [ -z "$FILES" ]; then
    exit 0
fi

# Determine executable
CODEI18N_CMD="codei18n"
if ! command -v $CODEI18N_CMD &> /dev/null; then
    echo "CodeI18n: command not found. Skipping."
    exit 0
fi

# Step 1: Update mappings for new comments
echo "CodeI18n: Updating mappings..."
$CODEI18N_CMD map update --scan-dir .

# Step 2: Translate any missing English translations
echo "CodeI18n: Translating missing entries..."
$CODEI18N_CMD translate --provider openai 2>/dev/null || echo "CodeI18n: Translation skipped (check API config)"

# Step 3: Convert staged files to English
for file in $FILES; do
    echo "CodeI18n: Processing $file..."
    $CODEI18N_CMD convert --to en --file "$file"
    
    # Add back to staging
    git add "$file"
done

echo "CodeI18n: Done."
exit 0
`

	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		log.Fatal("安装 hook 失败: %v", err)
	}

	log.Success("Hook 安装成功: %s", hookPath)
}

func runHookUninstall() {
	hookPath := filepath.Join(".git", "hooks", "pre-commit")
	if err := os.Remove(hookPath); err != nil {
		if os.IsNotExist(err) {
			log.Warn("Hook 不存在")
			return
		}
		log.Fatal("卸载 hook 失败: %v", err)
	}
	log.Success("Hook 已卸载")
}
