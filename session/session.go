package session

type (
	globalFlags struct {
		Verbose bool
		IsPipe  bool
	}
)

var flags globalFlags

func SetVerbose(verbose bool) {
	flags.Verbose = verbose
}

func Verbose() bool {
	return flags.Verbose
}

func SetIsPipe(pipe bool) {
	flags.IsPipe = pipe
}

func IsPipe() bool {
	return flags.IsPipe
}
