package components

import "dipt/internal/tui/theme"

// ASCII art logo
var logoLines = []string{
	`  ██████╗  ██╗ ██████╗ ████████╗`,
	`  ██╔══██╗ ██║ ██╔══██╗╚══██╔══╝`,
	`  ██║  ██║ ██║ ██████╔╝   ██║   `,
	`  ██║  ██║ ██║ ██╔═══╝    ██║   `,
	`  ██████╔╝ ██║ ██║        ██║   `,
	`  ╚═════╝  ╚═╝ ╚═╝        ╚═╝   `,
}

// RenderLogo 渲染渐变色 logo
func RenderLogo() string {
	result := ""
	for _, line := range logoLines {
		result += theme.GradientText(line) + "\n"
	}
	return result
}
