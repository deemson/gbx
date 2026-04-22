package exec

type Result struct {
	Args     []string
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}
