package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"youtube-channel-audio-downloader/internal/tui"
)

func Run(args []string) error {
	flags := flag.NewFlagSet("yt-channel-audio", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)

	targetURL := flags.String("url", "", "YouTube channel or single video URL")
	mode := flags.String("mode", "audio", "Download mode: audio, video, or both")
	output := flags.String("output", "./downloads", "Output directory")
	showHelp := flags.Bool("help", false, "Show help")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if *showHelp {
		printHelp(flags)
		return nil
	}
	outDir, err := filepath.Abs(*output)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return tui.Run(tui.Config{
		InitialURL:    strings.TrimSpace(*targetURL),
		InitialMode:   strings.TrimSpace(*mode),
		InitialOutput: outDir,
	})
}

func printHelp(flags *flag.FlagSet) {
	fmt.Println("Interactive YouTube downloader TUI (audio/video, channel or single-video URL).")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Printf("  %s [-url <youtube-url>] [-mode audio|video|both] [-output <directory>]\n", flags.Name())
	fmt.Println("")
	flags.PrintDefaults()
}
