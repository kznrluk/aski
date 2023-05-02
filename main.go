package main

import (
	"github.com/kznrluk/aski/lib"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "aski",
		Short: "aski is a very small and user-friendly ChatGPT client.",
		Long:  `aski is a very small and user-friendly ChatGPT client. It works hard to maintain context and establish communication.`,
		Run:   lib.Aski,
	}

	changeProfileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Select profile.",
		Long: "Profiles are usually located in the .aski/config.yaml file in the home directory." +
			"By using profiles, you can easily switch between different conversation contexts on the fly.",
		Run: lib.ChangeProfile,
	}

	rootCmd.AddCommand(changeProfileCmd)
	rootCmd.PersistentFlags().StringSliceP("file", "f", []string{}, "Input file(s) to start dialog from. Can be specified multiple times.")
	rootCmd.PersistentFlags().StringP("profile", "p", "", "Select the profile to use for this conversation, as defined in the .aski/config.yaml file.")
	rootCmd.PersistentFlags().StringP("content", "c", "", "Input text to start dialog from command line")
	rootCmd.PersistentFlags().StringP("restore", "r", "", "Restore conversations from history yaml files. Search pwd and .aski/history folders by default. Prefix match.")
	rootCmd.PersistentFlags().BoolP("rest", "", false, "When you specify this flag, you will communicate with the REST API instead of streaming. This can be useful if the communication is unstable or if you are not receiving responses properly.")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Debug logging")

	_ = rootCmd.Execute()
}
