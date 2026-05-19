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
	mode := flags.String("mode", "audio", "Download mode: audio or video")
	output := flags.String("output", "./downloads", "Output directory")
	ytDLPBin := flags.String("yt-dlp-bin", "yt-dlp", "Path to yt-dlp executable")
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
		InitialYTDLP:  strings.TrimSpace(*ytDLPBin),
	})
}

func printHelp(flags *flag.FlagSet) {
	fmt.Println("Interactive YouTube downloader TUI (audio/video, channel or single-video URL).")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Printf("  %s [-url <youtube-url>] [-mode audio|video] [-output <directory>] [-yt-dlp-bin <path>]\n", flags.Name())
	fmt.Println("")
	flags.PrintDefaults()
}
