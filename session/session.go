package session

type (
	globalFlags struct {
		Verbose bool
	}
)

var flags globalFlags

func SetVerbose(verbose bool) {
	flags.Verbose = verbose
}

func Verbose() bool {
	return flags.Verbose
}
